package sinks

import (
	"io"

	"example.com/core"
)

func TestSinks(s core.Source, writer io.Writer) {
	core.Sink([]interface{}{s}) // want "a source has reached a sink"
}
// makeScript
// 0 = {string} "<*ssa.Alloc> local example.com/core.Source (s)"
// 1 = {string} "<*ssa.Store> *t0 = s"
// 2 = {string} "<*ssa.Alloc> new [1]interface{} (slicelit)"
// 3 = {string} "<*ssa.IndexAddr> &t1[0:int]"
// 4 = {string} "<*ssa.UnOp> *t0"
// 5 = {string} "<*ssa.MakeInterface> make interface{} <- example.com/core.Source (t3)"
// 6 = {string} "<*ssa.Store> *t2 = t4"
// 7 = {string} "<*ssa.Slice> slice t1[:]"
// 8 = {string} "<*ssa.Alloc> new [1]interface{} (varargs)"
// 9 = {string} "<*ssa.IndexAddr> &t6[0:int]"
// 10 = {string} "<*ssa.MakeInterface> make interface{} <- []interface{} (t5)"
// 11 = {string} "<*ssa.Store> *t7 = t8"
// 12 = {string} "<*ssa.Slice> slice t6[:]"
// 13 = {string} "<*ssa.Call> example.com/core.Sink(t9...)"
// 14 = {string} "<*ssa.Return> return"
