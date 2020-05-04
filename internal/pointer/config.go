package pointer

import (
	"fmt"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/pointer"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

// callgraph.go uses ""==cwd and "testdata/src" as dir.
func getConfig(args []string, dir, gopath string) (*pointer.Config, error) {
	var pcnf = &packages.Config{
		Mode:  packages.LoadAllSyntax,
		Tests: false,
		Dir:   dir,
	}
	//if gopath != "" {
	//	pcnf.Env = append(os.Environ(), "GOPATH="+gopath)
	//}

	initial, err := packages.Load(pcnf, args...)
	if err != nil {
		return nil, err
	}
	if packages.PrintErrors(initial) > 0 {
		return nil, fmt.Errorf("packages contain errors")
	}

	// Create and build SSA-form program representation.
	prog, pkgs := ssautil.AllPackages(initial, 0)
	prog.Build()

	mains, err := mainPackages(pkgs)
	if err != nil {
		return nil, err
	}

	config := &pointer.Config{
		Mains:           mains,
		BuildCallGraph:  true,
		Reflection:      false,
		Queries:         nil,
		IndirectQueries: nil,
		Log:             nil,
	}

	return config, nil
}

// mainPackages returns the main packages to analyze.
// Each resulting package is named "main" and has a main function.
func mainPackages(pkgs []*ssa.Package) ([]*ssa.Package, error) {
	var mains []*ssa.Package
	for _, p := range pkgs {
		if p != nil && p.Pkg.Name() == "main" && p.Func("main") != nil {
			mains = append(mains, p)
		}
	}
	if len(mains) == 0 {
		return nil, fmt.Errorf("no main packages")
	}
	return mains, nil
}