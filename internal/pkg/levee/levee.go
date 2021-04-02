// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package internal

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"github.com/google/go-flow-levee/internal/pkg/config"
	"github.com/google/go-flow-levee/internal/pkg/fieldtags"
	"github.com/google/go-flow-levee/internal/pkg/propagation"
	"github.com/google/go-flow-levee/internal/pkg/source"
	"github.com/google/go-flow-levee/internal/pkg/suppression"
	"github.com/google/go-flow-levee/internal/pkg/utils"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/ssa"
)

var Analyzer = &analysis.Analyzer{
	Name:  "levee",
	Run:   run,
	Flags: config.FlagSet,
	Doc:   "reports attempts to source data to sinks",
	Requires: []*analysis.Analyzer{
		fieldtags.Analyzer,
		source.Analyzer,
		suppression.Analyzer,
	},
}

func run(pass *analysis.Pass) (interface{}, error) {
	conf, err := config.ReadConfig()
	if err != nil {
		return nil, err
	}
	funcSources := pass.ResultOf[source.Analyzer].(source.ResultType)
	taggedFields := pass.ResultOf[fieldtags.Analyzer].(fieldtags.ResultType)
	suppressedNodes := pass.ResultOf[suppression.Analyzer].(suppression.ResultType)

	sinks := identifySinks(funcSources, conf)
	callPropagators := identifyCallProps(funcSources)

	printSomeStats(pass, sinks, callPropagators, funcSources)

	for fn, sources := range funcSources {
		propagations := make(map[*source.Source]propagation.Propagation, len(sources))
		for _, s := range sources {
			propagations[s] = propagation.Taint(s.Node, conf, taggedFields)
		}

		for _, instr := range sinks[fn] {
			switch v := instr.(type) {
			case *ssa.Call:
				if callee := v.Call.StaticCallee(); callee != nil && conf.IsSink(utils.DecomposeFunction(callee)) {
					reportSourcesReachingSink(conf, pass, suppressedNodes, propagations, instr)
				}
			case *ssa.Panic:
				if conf.AllowPanicOnTaintedValues {
					continue
				}
				reportSourcesReachingSink(conf, pass, suppressedNodes, propagations, instr)
			}
		}
	}

	return nil, nil
}

type runInfo struct {
	pkg *types.Package

	srcs          source.ResultType
	sinks         map[*ssa.Function][]ssa.Instruction
	callprop      map[*ssa.Function][]ssa.Instruction
	intersec      []*ssa.Function
	sinkButNoProp []*ssa.Function
}

func (i runInfo) String(pass *analysis.Pass) string {
	nSrc, nSink, nProp, nIntersec := len(i.srcs), len(i.sinks), len(i.callprop), len(i.intersec)
	_ = nProp

	if nSrc == 0 || nSink <= nIntersec {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Package statistics (%v):\n", i.pkg)
	fmt.Fprintf(&b, "  Number of functions containing at least one source: %v\n", nSrc)
	fmt.Fprintf(&b, "    |- and at least one sink: %v\n", nSink)
	fmt.Fprintf(&b, "    |- BOTH a sink and a propagator: %v\n", nIntersec)
	fmt.Fprintf(&b, "    |- sink but no props: %v\n", len(i.sinkButNoProp))

	for _, fn := range i.sinkButNoProp {
		srcs := i.srcs[fn]

		fmt.Fprintf(&b, "        |- See functions:\n")
		fmt.Fprintf(&b, "            - %v:\n", fn)
		fmt.Fprintf(&b, "            - %v\n", pass.Fset.Position(fn.Pos()))
		for _, s := range srcs {
			fmt.Fprintf(&b, "            - source %v\n", s)
			fmt.Fprintf(&b, "            - %v\n", pass.Fset.Position(s.Pos()))
		}
	}

	return b.String()
}

func printSomeStats(pass *analysis.Pass, sinks map[*ssa.Function][]ssa.Instruction, callPropagators map[*ssa.Function][]ssa.Instruction, funcSources source.ResultType) {
	info := newRunInfo(pass, sinks, callPropagators, funcSources)

	if s := info.String(pass); s != "" {
		fmt.Println(s)
	}
}

func newRunInfo(pass *analysis.Pass, sinks map[*ssa.Function][]ssa.Instruction, callPropagators map[*ssa.Function][]ssa.Instruction, funcSources source.ResultType) runInfo {
	var intersection []*ssa.Function
	var sinkButNoProp []*ssa.Function
	for fn, _ := range sinks {
		if _, ok := callPropagators[fn]; ok {
			intersection = append(intersection, fn)
		} else {
			sinkButNoProp = append(sinkButNoProp, fn)
		}
	}

	info := runInfo{
		pkg:           pass.Pkg,
		srcs:          funcSources,
		sinks:         sinks,
		callprop:      callPropagators,
		intersec:      intersection,
		sinkButNoProp: sinkButNoProp,
	}
	return info
}

// identifySinks returns a map of function to sink calls within that function,
// restricted to those functions which have sources present
func identifySinks(funcSources source.ResultType, conf *config.Config) map[*ssa.Function][]ssa.Instruction {
	var sinks = make(map[*ssa.Function][]ssa.Instruction)

	for fn, _ := range funcSources {
		for _, b := range fn.Blocks {
			for _, instr := range b.Instrs {
				switch v := instr.(type) {
				case *ssa.Call:
					if callee := v.Call.StaticCallee(); callee != nil && conf.IsSink(utils.DecomposeFunction(callee)) {
						sinks[fn] = append(sinks[fn], instr)
					}
				case *ssa.Panic:
					if conf.AllowPanicOnTaintedValues {
						continue
					}
					sinks[fn] = append(sinks[fn], instr)
				}
			}
		}
	}
	return sinks
}

// identifyCallProps returns a map of function to stdlib propagator calls within that function,
// restricted to those functions which have sources present
func identifyCallProps(funcSources source.ResultType) map[*ssa.Function][]ssa.Instruction {
	var callPropagators = make(map[*ssa.Function][]ssa.Instruction)

	for fn, _ := range funcSources {
		for _, b := range fn.Blocks {
			for _, instr := range b.Instrs {
				switch v := instr.(type) {
				case *ssa.Call:
					if propagation.HasInterfaceSummary(v) || propagation.HasStaticSummary(v) {
						callPropagators[fn] = append(callPropagators[fn], v)
					}
				}
			}
		}
	}
	return callPropagators
}

func reportSourcesReachingSink(conf *config.Config, pass *analysis.Pass, suppressedNodes suppression.ResultType, propagations map[*source.Source]propagation.Propagation, sink ssa.Instruction) {
	for src, prop := range propagations {
		if prop.IsTainted(sink) && !isSuppressed(sink.Pos(), suppressedNodes, pass) {
			report(conf, pass, src, sink.(ssa.Node))
			break
		}
	}
}

func isSuppressed(pos token.Pos, suppressedNodes suppression.ResultType, pass *analysis.Pass) bool {
	for _, f := range pass.Files {
		if pos < f.Pos() || f.End() < pos {
			continue
		}
		// astutil.PathEnclosingInterval produces the list of nodes that enclose the provided
		// position, from the leaf node that directly contains it up to the ast.File node
		path, _ := astutil.PathEnclosingInterval(f, pos, pos)
		if len(path) < 2 {
			return false
		}
		// Given the position of a call, path[0] holds the ast.CallExpr and
		// path[1] holds the ast.ExprStmt. A suppressing comment may be associated
		// with the name of the function being called (Ident, SelectorExpr), with the
		// call itself (CallExpr), or with the entire expression (ExprStmt).
		if ce, ok := path[0].(*ast.CallExpr); ok {
			switch t := ce.Fun.(type) {
			case *ast.Ident:
				/*
					Sink( // levee.DoNotReport
				*/
				if suppressedNodes.IsSuppressed(t) {
					return true
				}
			case *ast.SelectorExpr:
				/*
					core.Sink( // levee.DoNotReport
				*/
				if suppressedNodes.IsSuppressed(t.Sel) {
					return true
				}
			}
		} else {
			fmt.Printf("unexpected node received: %v (type %T); please report this issue\n", path[0], path[0])
		}
		return suppressedNodes.IsSuppressed(path[0]) || suppressedNodes.IsSuppressed(path[1])
	}
	return false
}

func report(conf *config.Config, pass *analysis.Pass, source *source.Source, sink ssa.Node) {
	var b strings.Builder
	b.WriteString("a source has reached a sink")
	fmt.Fprintf(&b, "\n source: %v", pass.Fset.Position(source.Pos()))
	if conf.ReportMessage != "" {
		fmt.Fprintf(&b, "\n %v", conf.ReportMessage)
	}
	pass.Reportf(sink.Pos(), b.String())
}
