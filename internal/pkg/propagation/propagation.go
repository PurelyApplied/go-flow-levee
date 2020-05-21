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

package propagation

import (
	"reflect"

	"github.com/google/go-flow-levee/internal/pkg/config"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ssa"

	"github.com/google/go-flow-levee/internal/pkg/source"
)

var Analyzer = &analysis.Analyzer{
	Name:       "levee",
	Doc:        "reports attempts to source data to sinks",
	Flags:      config.FlagSet,
	Run:        run,
	Requires:   []*analysis.Analyzer{source.Analyzer},
	ResultType: reflect.TypeOf(map[*ssa.Function][]*source.Source{}),
}

func run(pass *analysis.Pass) (interface{}, error) {
	conf, err := config.ReadConfig()
	if err != nil {
		return nil, err
	}

	sourcesMap := pass.ResultOf[source.Analyzer].(source.ResultType)
	// TODO Distinguish propagated taint from source
	propagationMap := make(map[*ssa.Function][]*source.Source)
	for fn := range sourcesMap {
		var taints []*source.Source
		for _, b := range fn.Blocks {
			for _, instr := range b.Instrs {
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
						if a := sendsToIOWriter(conf, v); a != nil {
							taints = append(taints, source.New(a, conf))
						} else {
							taints = append(taints, source.New(v, conf))
						}
					}
				}
			}
		}

		if len(taints) > 0 {
			propagationMap[fn] = taints
		}
	}

	for _, taints := range propagationMap {
		for _, t := range taints {
			pass.Reportf(t.Node().Pos(), "this value becomes tainted")
		}
	}

	return propagationMap, nil
}

func sendsToIOWriter(c *config.Config, call *ssa.Call) ssa.Node {
	if call.Call.Signature().Params().Len() == 0 {
		return nil
	}

	firstArg := call.Call.Signature().Params().At(0)
	if c.PropagatorArgs.ArgumentTypeRE.MatchString(firstArg.Type().String()) {
		if a, ok := call.Call.Args[0].(*ssa.MakeInterface); ok {
			return a.X.(ssa.Node)
		}
	}

	return nil
}
