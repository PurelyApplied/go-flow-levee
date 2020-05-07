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

package sources

import (
	"go/types"
	"reflect"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/ssa"
	"google.com/go-flow-levee/internal/common"

	"github.com/eapache/queue"
)

type SourceMap map[*ssa.Function][]*Source

func (res SourceMap) add(fn *ssa.Function, sources []*Source) {
	if len(sources) == 0 {
		return
	}

	res[fn] = sources
}

// MakeReport issues findings as analysis report.
var MakeReport bool

var Analyzer = &analysis.Analyzer{
	Name:       "sources",
	Doc:        "reports attempts to source data to sinks",
	Flags:      common.SharedFlags,
	Run:        run,
	Requires:   []*analysis.Analyzer{buildssa.Analyzer, common.ConfigLoader},
	ResultType: reflect.TypeOf(new(SourceMap)).Elem(),
}

// source represents a source in an SSA call tree.
// It is based on ssa.Node, with the added functionality of computing the recursive graph of
// its referrers.
// source.sanitized notes sanitizer calls that sanitize this source
type Source struct {
	Node       ssa.Node
	marked     map[ssa.Node]bool
	sanitizers []*sanitizer
	Kind
}

type Kind string

const (
	freeVarKind Kind = "freeVarKind"
	paramKind   Kind = "param"
	fieldAddrKind Kind = "fieldAddr"
	allocKind Kind = "alloc"
)

func newSource(in ssa.Node, config *common.Config, kind Kind) *Source {
	a := &Source{
		Node:   in,
		marked: make(map[ssa.Node]bool),
		Kind: kind,
	}
	a.bfs(config)
	return a
}

// bfs performs Breadth-First-Search on the def-use graph of the input source.
// While traversing the graph we also look for potential sanitizers of this source.
// If the source passes through a sanitizer, bfs does not continue through that Node.
func (a *Source) bfs(config *common.Config) {
	q := queue.New()
	q.Add(a.Node)
	a.marked[a.Node] = true

	for q.Length() > 0 {
		e := q.Remove().(ssa.Node)

		if c, ok := e.(*ssa.Call); ok && config.IsSanitizer(c) {
			a.sanitizers = append(a.sanitizers, &sanitizer{call: c})
			continue
		}

		if e.Referrers() == nil {
			continue
		}

		for _, r := range *e.Referrers() {
			if _, ok := a.marked[r.(ssa.Node)]; ok {
				continue
			}
			a.marked[r.(ssa.Node)] = true

			// Need to stay within the scope of the function under analysis.
			if call, ok := r.(*ssa.Call); ok && !config.IsPropagator(call) {
				continue
			}

			// Do not follow innocuous field access (e.g. Cluster.Zone)
			if addr, ok := r.(*ssa.FieldAddr); ok && !config.IsSourceFieldAddr(addr) {
				continue
			}

			q.Add(r)
		}
	}
}

// HasPathTo returns true when a Node is part of declaration-use graph.
func (a *Source) HasPathTo(n ssa.Node) bool {
	return a.marked[n]
}

type sanitizer struct {
	call *ssa.Call
}

func run(pass *analysis.Pass) (interface{}, error) {
	conf := pass.ResultOf[common.ConfigLoader].(*common.Config)
	ssaInput := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)

	result := make(SourceMap)

	for _, fn := range ssaInput.SrcFuncs {
		var sources []*Source

		// Sources can come from parameters, closure, or result from instructions in blocks
		for _, p := range fn.Params {
			if s, ok := sourceFromParameter(p, conf); ok {
				sources = append(sources, s)
			}
		}

		for _, fv := range fn.FreeVars {
			if s, ok := sourceFromFreeVar(fv, conf); ok {
				sources = append(sources, s)
			}
		}

		for _, bl := range fn.Blocks {
			if bl == fn.Recover {
				// TODO Handle calls to log in a recovery block.
				continue
			}

			for _, instr := range bl.Instrs {
				switch v := instr.(type) {
				case *ssa.Alloc:
					if conf.IsSource(common.DereferenceRecursive(v.Type())) {
						sources = append(sources, newSource(v, conf, allocKind))
					}

					// Handling the case where PII may be in a receiver
					// (ex. func(b *something) { log.Info(something.PII) }
				case *ssa.FieldAddr:
					if conf.IsSource(common.DereferenceRecursive(v.Type())) {
						sources = append(sources, newSource(v, conf, fieldAddrKind))
					}

				case *ssa.Field:
					// TODO
				}
			}
		}
		result.add(fn, sources)
	}

	if MakeReport {
		for _, srcs := range result {
			for _, s := range srcs {
				pass.Reportf(s.Node.Pos(), "source identified: %v", s.Kind)

			}
		}
	}
	return result, nil
}

func sourceFromParameter(p *ssa.Parameter, conf *common.Config) (*Source, bool) {
	deref := common.DereferenceRecursive(p.Type())
	if n, ok := deref.(*types.Named); ok && conf.IsSource(n) {
		return newSource(p, conf, paramKind), true
	}
	return nil, false
}

func sourceFromFreeVar(fv *ssa.FreeVar, conf *common.Config) (*Source, bool) {
	deref := common.DereferenceRecursive(fv.Type())
	if n, ok := deref.(*types.Named); ok && conf.IsSource(n) {
		return newSource(fv, conf, freeVarKind), true
	}
	return nil, false
}
