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

package source

import (
	"fmt"
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/ssa"
)

var testAnalyzer = &analysis.Analyzer{
	Name:     "source",
	Run:      run,
	Doc:      "test harness for the logic related to sources",
	Requires: []*analysis.Analyzer{buildssa.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	in := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)
	var calls []*ssa.Call
	for _, fn := range in.SrcFuncs {
		for _, b := range fn.Blocks {
			for _, i := range b.Instrs {
				fmt.Printf("%v : %T\n", i, i)
			}
		}
	}

	return calls, nil
}

func TestDomination(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, testAnalyzer)
}
