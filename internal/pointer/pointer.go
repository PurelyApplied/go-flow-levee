// This has copied callgraph/main.go wholesale while I figure some things out.
package pointer

import (
	"flag"
	"fmt"

	"golang.org/x/tools/go/analysis"
)

type Config struct {
	Args []string
	Dir string
}

var basicConf Config

var Analyzer = &analysis.Analyzer{
	Name:             "",
	Doc:              "",
	Flags:            flag.FlagSet{},
	Run:              run,
	RunDespiteErrors: false,
	Requires:         nil,
	ResultType:       nil,
	FactTypes:        nil,
}

func run(pass *analysis.Pass) (interface{}, error) {
	fmt.Println("Arrived in pass for pkg", pass.Pkg)
	config, err := getConfig(basicConf.Args, basicConf.Dir, basicConf.Dir)
	if err != nil {
		return nil, err
	}

	_ = config
	return nil, nil
}
