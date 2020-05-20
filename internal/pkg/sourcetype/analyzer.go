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
	localFacts  []*DeclarationFact
	importFacts map[*types.Package]*PackageSourceFacts
}

func (sc *sourceClassifier) ExportFacts(pass *analysis.Pass) {
	// Only export this package's facts
	pass.ExportPackageFact(&PackageSourceFacts{
		pkg:     pass.Pkg,
		sources: sc.localFacts,
	})

	for _, declFact := range sc.localFacts {
		pass.ExportObjectFact(declFact.typ, declFact)
		for _, fldFact := range declFact.fieldFacts {
			// Type aliasing can lean to the locally declared source referring to out-of-package fields.
			// Exporting outside of the pass's package is prohibited.
			if fldFact.field.Pkg() == pass.Pkg {
				pass.ExportObjectFact(fldFact.field, fldFact)
			}
		}
	}
}

func (sc *sourceClassifier) IsSource(t types.Type) bool {
	for _, declFact := range sc.localFacts {
		if t == declFact.typ.Type() {
			return true
		}
	}

	for _, pkgFacts := range sc.importFacts {
		for _, declFact := range pkgFacts.sources {
			if t == declFact.typ.Type() {
				return true
			}
		}
	}

	return false
}

func (sc *sourceClassifier) IsSourceField(t types.Type, f *types.Var) bool {
	for _, declFact := range sc.localFacts {
		if t == declFact.typ.Type() {
			for _, fld := range declFact.fieldFacts {
				if fld.field == f {
					return true
				}
			}
		}
	}

	for _, pkgFacts := range sc.importFacts {
		for _, declFact := range pkgFacts.sources {
			if t == declFact.typ.Type() {
				for _, fld := range declFact.fieldFacts {
					if fld.field == f {
						return true
					}
				}
			}
		}
	}

	return false
}

func newSourceClassifier() *sourceClassifier {
	return &sourceClassifier{
		importFacts: make(map[*types.Package]*PackageSourceFacts),
	}
}

type PackageSourceFacts struct {
	pkg     *types.Package
	sources []*DeclarationFact
}

func (p PackageSourceFacts) AFact() {}
func (p PackageSourceFacts) String() string {
	return fmt.Sprintf("sources declared: %d", len(p.sources))
}

type FieldDeclarationFact struct {
	field *types.Var
}

func (f FieldDeclarationFact) AFact() {

}

func (f FieldDeclarationFact) String() string {
	return "source field declaration"
}

// DeclarationFact tracks source type declarations across package boundaries.
type DeclarationFact struct {
	typ        types.Object
	fieldFacts []*FieldDeclarationFact
}

func (s DeclarationFact) String() string {
	return "source type declaration"
}

func (s DeclarationFact) AFact() {}

var Analyzer = &analysis.Analyzer{
	Name:       "sourcetypes",
	Doc:        "This analyzer identifies types.Types values which contain dataflow sources.",
	Flags:      config.FlagSet,
	Run:        run,
	Requires:   []*analysis.Analyzer{buildssa.Analyzer},
	ResultType: reflect.TypeOf(new(sourceClassifier)),
	FactTypes:  []analysis.Fact{new(PackageSourceFacts), new(FieldDeclarationFact), new(FieldDeclarationFact)},
}

func run(pass *analysis.Pass) (interface{}, error) {
	ssaInput := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)
	conf, err := config.ReadConfig()
	if err != nil {
		return nil, err
	}

	sc := newSourceClassifier()
	// Load source identification from imported packages.
	for _, pkg := range ssaInput.Pkg.Pkg.Imports() {
		var fct PackageSourceFacts
		pass.ImportPackageFact(pkg, &fct)
		if len(fct.sources) > 0 {
			sc.importFacts[pkg] = &fct
		}
	}

	// Members contains all named entities
	for _, mem := range ssaInput.Pkg.Members {
		if ssaType, ok := mem.(*ssa.Type); ok {
			if conf.IsSource(ssaType.Type()) {
				// If the member names a struct declaration, examine the fields for additional tagging.
				var sourceFields []*FieldDeclarationFact
				if under, ok := ssaType.Type().Underlying().(*types.Struct); ok {
					for i := 0; i < under.NumFields(); i++ {
						fld := under.Field(i)
						if conf.IsSourceField(ssaType.Type(), fld) {
							sourceFields = append(sourceFields, newFieldFact(fld))
						}
					}
				}
				// TODO Warn if no fields are identified

				srcFact := newTypeDeclarationFact(ssaType.Object(), sourceFields)
				sc.localFacts = append(sc.localFacts, srcFact)
			}
		}
	}

	sc.ExportFacts(pass)
	return sc, nil
}

func newTypeDeclarationFact(typ types.Object, flds []*FieldDeclarationFact) *DeclarationFact {
	return &DeclarationFact{
		typ:        typ,
		fieldFacts: flds,
	}
}

func newFieldFact(fld *types.Var) *FieldDeclarationFact {
	return &FieldDeclarationFact{field: fld}
}
