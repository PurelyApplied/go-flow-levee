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
	"math"
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
}

func newSource(in ssa.Node, config *common.Config) *Source {
	a := &Source{
		Node:   in,
		marked: make(map[ssa.Node]bool),
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

func (a *Source) IsSanitizedAt(call ssa.Instruction) bool {
	for _, s := range a.sanitizers {
		if s.dominates(call) {
			return true
		}
	}

	return false
}

type sanitizer struct {
	call *ssa.Call
}

// dominates returns true if the sanitizer dominates the supplied instruction.
// In the context of SSA, domination implies that
// if instructions X executes and X dominates Y, then Y is guaranteed to execute and to be
// executed after X.
func (s sanitizer) dominates(target ssa.Instruction) bool {
	if s.call.Parent() != target.Parent() {
		// Instructions are in different functions.
		return false
	}

	if !s.call.Block().Dominates(target.Block()) {
		return false
	}

	if s.call.Block() == target.Block() {
		parentIdx := math.MaxInt64
		childIdx := 0
		for i, instr := range s.call.Block().Instrs {
			if instr == s.call {
				parentIdx = i
			}
			if instr == target {
				childIdx = i
				break
			}
		}
		return parentIdx < childIdx
	}

	for _, d := range s.call.Block().Dominees() {
		if target.Block() == d {
			return true
		}
	}

	return false
}

// varargs represents a variable length argument.
// Concretely, it abstract over the fact that the varargs internally are represented by an ssa.Slice
// which contains the underlying values of for vararg members.
// Since many sink functions (ex. log.Info, fmt.Errorf) take a vararg argument, being able to
// get the underlying values of the vararg members is important for this analyzer.
type varargs struct {
	slice   *ssa.Slice
	sources []*Source
}

// newVarargs constructs varargs. SSA represents varargs as an ssa.Slice.
func newVarargs(s *ssa.Slice, sources []*Source) *varargs {
	a, ok := s.X.(*ssa.Alloc)
	if !ok || a.Comment != "varargs" {
		return nil
	}
	var (
		referredSources []*Source
	)

	for _, r := range *a.Referrers() {
		idx, ok := r.(*ssa.IndexAddr)
		if !ok {
			continue
		}

		if idx.Referrers() != nil && len(*idx.Referrers()) != 1 {
			continue
		}

		// IndexAddr and Store instructions are inherently linked together.
		// IndexAddr returns an address of an element within a Slice, which is followed by
		// a Store instructions to place a value into the address provided by IndexAddr.
		store := (*idx.Referrers())[0].(*ssa.Store)

		for _, s := range sources {
			if s.HasPathTo(store.Val.(ssa.Node)) {
				referredSources = append(referredSources, s)
				break
			}
		}
	}

	return &varargs{
		slice:   s,
		sources: referredSources,
	}
}

func (v *varargs) referredByCallWithPattern(patterns []common.NameMatcher) *ssa.Call {
	if v.slice.Referrers() == nil || len(*v.slice.Referrers()) != 1 {
		return nil
	}

	c, ok := (*v.slice.Referrers())[0].(*ssa.Call)
	if !ok || c.Call.StaticCallee() == nil {
		return nil
	}

	for _, p := range patterns {
		if p.MatchMethodName(c) {
			return c
		}
	}

	return nil
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
						sources = append(sources, newSource(v, conf))
					}

					// Handling the case where PII may be in a receiver
					// (ex. func(b *something) { log.Info(something.PII) }
				case *ssa.FieldAddr:
					if conf.IsSource(common.DereferenceRecursive(v.Type())) {
						sources = append(sources, newSource(v, conf))
					}

				case *ssa.Field:
					// TODO
				}
			}
		}
		result.add(fn, sources)
	}

	return result, nil
}

func sourceFromParameter(p *ssa.Parameter, conf *common.Config) (*Source, bool) {
	// TODO Refine this detection.
	if t, ok := p.Type().(*types.Pointer); ok {
		if n, ok := t.Elem().(*types.Named); ok && conf.IsSource(n) {
			return newSource(p, conf), true
		}
	}
	return nil, false
}

func sourceFromFreeVar(fv *ssa.FreeVar, conf *common.Config) (*Source, bool) {
	// TODO Refine this detection.
	if t, ok := fv.Type().(*types.Pointer); ok {
		if s, ok := common.DereferenceRecursive(t).(*types.Named); ok && conf.IsSource(s) {
			return newSource(fv, conf), true
		}
	}
	return nil, false
}
