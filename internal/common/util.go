package common

import "go/types"

func DereferenceRecursive(t types.Type) types.Type {
	for {
		tt, ok := t.Underlying().(*types.Pointer)
		if !ok {
			return t
		}
		t = tt.Elem()
	}
}

func IsTestPkg(p *types.Package) bool {
	for _, im := range p.Imports() {
		if im.Name() == "testing" {
			return true
		}
	}
	return false
}
