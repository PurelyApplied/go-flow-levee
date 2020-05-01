package common

import (
	"reflect"
	"sync"

	"golang.org/x/tools/go/analysis"
)

var ConfigLoader = &analysis.Analyzer{
	Name:       "commonConfig",
	Doc:        "loads shared configuration",
	Flags:      SharedFlags,
	Run:        run,
	ResultType: reflect.TypeOf(new(Config)),
}

var conf *Config
var err error
var loadOnce sync.Once

func run(_ *analysis.Pass) (interface{}, error) {
	loadOnce.Do(func() {
		conf, err = readConfig(configFile)
	})
	return conf, err
}
