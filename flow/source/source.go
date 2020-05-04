// package source is a utility analyzer to identify struct fields of interest.
package source

import (
	"flag"
	"fmt"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/ssa"
)

var Analyzer = &analysis.Analyzer{
	Name:     "source",
	Doc:      "A utility to identify struct fields of interest",
	Flags:    flag.FlagSet{},
	Run:      run,
	Requires: []*analysis.Analyzer{buildssa.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	ssaInput := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)

	var reports []string
	for name, mem := range ssaInput.Pkg.Members {
		if ssaType, ok := mem.(*ssa.Type); ok {
			typ := ssaType.Type()
			if base, ok := getBaseStruct(typ); ok {
				info, interesting := getInfo(base)
				if interesting || name == "Secret" {
					reports = append(reports, fmt.Sprintf("%v {\n  %v\n}\n", name, strings.Join(info, "\n  ")))
				}
			}
		}
	}

	if len(reports) > 0 {
		fmt.Printf("Reporting for %v:\n%v", pass.Pkg, strings.Join(reports, "\n\n"))
	}
	return nil, nil
}

func getBaseStruct(named types.Type) (*types.Struct, bool) {
	under := named.Underlying()
	for {
		switch x := under.(type) {
		case *types.Struct:
			return x, true
		case *types.Named:
			return getBaseStruct(x.Underlying())

		case *types.Array:
			return getBaseStruct(x.Elem())
		case *types.Slice:
			return getBaseStruct(x.Elem())
		case *types.Pointer:
			return getBaseStruct(x.Elem())

			// TODO Tuple, map, chan?
		case *types.Tuple:
			return nil, false
		case *types.Map:
			return nil, false
		case *types.Chan:
			return nil, false

		case *types.Basic:
			return nil, false
		case *types.Signature:
			return nil, false
		case *types.Interface:
			return nil, false
		}
	}
}

func getInfo(struc *types.Struct) ([]string, bool) {
	var interesting bool
	var info []string
	for i := 0; i < struc.NumFields(); i++ {
		f := struc.Field(i)
		tags := struc.Tag(i)
		interesting = interesting || strings.Contains(tags, "protobuf")
		info = append(info, fmt.Sprintf("%v `%v`", f, tags))
	}
	return info, interesting
}
