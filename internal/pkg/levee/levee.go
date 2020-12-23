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
	"strings"

	"github.com/google/go-flow-levee/internal/pkg/config"
	"github.com/google/go-flow-levee/internal/pkg/fieldtags"
	"github.com/google/go-flow-levee/internal/pkg/levee/propagation"
	"github.com/google/go-flow-levee/internal/pkg/utils"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ssa"

	"github.com/google/go-flow-levee/internal/pkg/source"
)

var Analyzer = &analysis.Analyzer{
	Name:     "levee",
	Run:      run,
	Flags:    config.FlagSet,
	Doc:      "reports attempts to source data to sinks",
	Requires: []*analysis.Analyzer{source.Analyzer, fieldtags.Analyzer},
}

type reportItems struct {
	s *source.Source
	v *ssa.Call
}

func run(pass *analysis.Pass) (interface{}, error) {
	conf, err := config.ReadConfig()
	if err != nil {
		return nil, err
	}
	funcSources := pass.ResultOf[source.Analyzer].(source.ResultType)
	taggedFields := pass.ResultOf[fieldtags.Analyzer].(fieldtags.ResultType)

	var items []reportItems
	for fn, sources := range funcSources {
		propagations := make(map[*source.Source]propagation.Propagation, len(sources))
		for _, s := range sources {
			propagations[s] = propagation.Dfs(s.Node, conf, taggedFields)
		}

		for _, b := range fn.Blocks {
			for _, instr := range b.Instrs {
				switch v := instr.(type) {

				case *ssa.Call:
					if callee := v.Call.StaticCallee(); callee != nil && conf.IsSink(utils.DecomposeFunction(callee)) {
						// build items
						reportSourcesReachingSink(pass, propagations, instr)
					}

				case *ssa.Panic:
					if conf.AllowPanicOnTaintedValues {
						continue
					}
					// build items
					reportSourcesReachingSink(pass, propagations, instr)
				}
			}
		}
	}

	reportAllItems(pass, conf, items)
	return nil, nil
}

func reportSourcesReachingSink(pass *analysis.Pass, propagations map[*source.Source]propagation.Propagation, sink ssa.Instruction) {
	for source, prop := range propagations {
		if prop.IsTainted(sink) {
			report(pass, source, sink.(ssa.Node))
			break
		}
	}
}

type ErrMessageFact string

func (_ ErrMessageFact) AFact() {}

func (emf ErrMessageFact) String() string {
	return string(emf)
}

func reportAllItems(pass *analysis.Pass, conf *config.Config, items []reportItems) {
	if len(items) > 0 && conf.MessageOnError != "" {
		fact := ErrMessageFact(conf.MessageOnError)
		pass.ExportPackageFact(&fact)
	}

	for _, item := range items {
		report(pass, item.s, item.v)
	}
}

func report(pass *analysis.Pass, source *source.Source, sink ssa.Node) {
	var b strings.Builder
	b.WriteString("a source has reached a sink")
	fmt.Fprintf(&b, ", source: %v", pass.Fset.Position(source.Pos()))
	pass.Reportf(sink.Pos(), b.String())
}
