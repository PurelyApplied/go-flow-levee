// Copyright 2020 Google LLC
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

package sourcetype

import (
	"fmt"
	"go/types"
	"reflect"

	"github.com/google/go-flow-levee/internal/pkg/config"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/ssa"
)

type sourceClassifier struct {
	// ssaTypes maps the package ssa.Member to the corresponding types.Type
	// This may be many-to-one if the Member is a type alias
	ssaTypes map[*ssa.Type]types.Type

	// sources maps the source type to its respective source fields.
	sources map[types.Type][]*types.Var

	facts []*DeclarationFact
}

func (sc *sourceClassifier) ExportFacts(pass *analysis.Pass) {
	for _, f := range sc.facts {
		pass.ExportObjectFact(f.Object, f)
	}
}

func (sc *sourceClassifier) IsSource(t types.Type) bool {
	_, contained := sc.sources[t]
	return contained
}

func (sc *sourceClassifier) IsSourceField(t types.Type, f *types.Var) bool {
	for _, fld := range sc.sources[t] {
		if fld == f {
			return true
		}
	}
	return false
}

func (sc *sourceClassifier) add(ssaType *ssa.Type, sourceFields []*types.Var) {
	sc.ssaTypes[ssaType] = ssaType.Type()
	sc.sources[ssaType.Type()] = sourceFields
	sc.facts = append(sc.facts, &DeclarationFact{ssaType.Object()})
}

func (sc *sourceClassifier) makeReport(pass *analysis.Pass) {
	// Alias types can double-report field identification.
	// Don't report the same underlying types.Type more than once.
	alreadyReported := make(map[types.Type]bool)

	sc.ExportFacts(pass)

	for _, typ := range sc.ssaTypes {
		if !alreadyReported[typ] {
			alreadyReported[typ] = true
			for _, f := range sc.sources[typ] {
				// Only report on your pass's package to avoid multiple reportings of cross-package alias or named types
				if f.Pkg() != pass.Pkg {
					continue
				}
				pass.Reportf(f.Pos(), "source field declaration identified: %v (from %v)", f, pass.Pkg)
			}

		}
	}
}

func newSourceClassifier() *sourceClassifier {
	return &sourceClassifier{
		sources:  make(map[types.Type][]*types.Var),
		ssaTypes: make(map[*ssa.Type]types.Type),
	}
}

// DeclarationFact tracks source type declarations across package boundaries.
type DeclarationFact struct {
	types.Object
}

func (s DeclarationFact) String() string {
	return fmt.Sprintf("source type declaration: %v", s.Object)
}

func (s DeclarationFact) AFact() {}

var Analyzer = &analysis.Analyzer{
	Name:       "sourcetypes",
	Doc:        "This analyzer identifies types.Types values which contain dataflow sources.",
	Flags:      config.FlagSet,
	Run:        run,
	Requires:   []*analysis.Analyzer{buildssa.Analyzer},
	ResultType: reflect.TypeOf(new(sourceClassifier)),
	FactTypes:  []analysis.Fact{new(DeclarationFact)},
}

func run(pass *analysis.Pass) (interface{}, error) {
	ssaInput := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)
	conf, err := config.ReadConfig()
	if err != nil {
		return nil, err
	}

	sc := newSourceClassifier()
	// Members contains all named entities
	for _, mem := range ssaInput.Pkg.Members {
		if ssaType, ok := mem.(*ssa.Type); ok {
			if conf.IsSource(ssaType.Type()) {
				var sourceFields []*types.Var
				// If the member names a struct declaration, examine the fields for additional tagging.
				if under, ok := ssaType.Type().Underlying().(*types.Struct); ok {
					for i := 0; i < under.NumFields(); i++ {
						fld := under.Field(i)
						if conf.IsSourceField(ssaType.Type(), fld) {
							sourceFields = append(sourceFields, fld)
						}
					}
				}

				sc.add(ssaType, sourceFields)
			}
		}
	}

	if config.Reporting.Has("sourcetype") {
		sc.makeReport(pass)
	}

	return sc, nil
}
