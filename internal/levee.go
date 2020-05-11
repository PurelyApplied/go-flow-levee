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

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/ssa"
	"google.com/go-flow-levee/internal/pkg/config"
	"google.com/go-flow-levee/internal/pkg/source"
)

var configFile string

func init() {
	Analyzer.Flags.StringVar(&configFile, "config", "config.json", "path to analysis configuration file")
}

var Analyzer = &analysis.Analyzer{
	Name:     "levee",
	Run:      run,
	Doc:      "reports attempts to source data to sinks",
	Requires: []*analysis.Analyzer{buildssa.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	conf, err := config.ReadConfig(configFile)
	if err != nil {
		return nil, err
	}
	// TODO: respect configuration scope

	ssaInput := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)

	sourcesMap := source.IdentifySources(conf, ssaInput)

	// Only examine functions that have sources
	for fn, sources := range sourcesMap {
		//log.V(2).Infof("Processing function %v", fn)

		for _, b := range fn.Blocks {
			if b == fn.Recover {
				// TODO Handle calls to sinks in a recovery block.
				continue // skipping Recover since it does not have instructions, rather a single block.
			}

			for _, instr := range b.Instrs {
				//log.V(2).Infof("Inst: %v %T", instr, instr)
				switch v := instr.(type) {
				case *ssa.Call:
					switch {
					case conf.IsPropagator(v):
						// Handling the case where sources are propagated to io.Writer
						// (ex. proto.MarshalText(&buf, c)
						// In such cases, "buf" becomes a source, and not the return value of the propagator.
						// TODO Do not hard-code logging sinks usecase
						// TODO  Handle case of os.Stdout and os.Stderr.
						// TODO  Do not hard-code the position of the argument, instead declaratively
						//  specify the position of the propagated source.
						// TODO  Consider turning propagators that take io.Writer into sinks.
						if a := conf.SendsToIOWriter(v); a != nil {
							sources = append(sources, source.NewSource(a, conf))
						} else {
							//log.V(2).Infof("Adding source: %v %T", v.Value(), v.Value())
							sources = append(sources, source.NewSource(v, conf))
						}

					case conf.IsSink(v):
						// TODO Only variadic sink arguments are currently detected.
						if v.Call.Signature().Variadic() && len(v.Call.Args) > 0 {
							lastArg := v.Call.Args[len(v.Call.Args)-1]
							if varargs, ok := lastArg.(*ssa.Slice); ok {
								if sinkVarargs := source.NewVarargs(varargs, sources); sinkVarargs != nil {
									for _, s := range sinkVarargs.Sources {
										if !s.IsSanitizedAt(v) {
											report(pass, s, v)
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return nil, nil
}

func report(pass *analysis.Pass, s *source.Source, sink ssa.Node) {
	var b strings.Builder
	b.WriteString("a source has reached a sink")
	fmt.Fprintf(&b, ", source: %v", pass.Fset.Position(s.Node.Pos()))
	pass.Reportf(sink.Pos(), b.String())
}
