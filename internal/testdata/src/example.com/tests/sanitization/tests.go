package sinks

import (
	"example.com/core"
	//
	//"google.com/go-flow-levee/internal/testdata/src/example.com/core"
)

func TestGraph(s core.Source) {
	format := "This is a format string.  %v"
	s2 := s
	core.OneArgSink(s2)  // want "a source has reached a sink"
	core.Sink(format, s) // want "a source has reached a sink"
}

//func TestGraph(s core.Source) {
//	sp := &s
//	s2 := dereference(sp)
//	s3 := toInterface(s2)
//	core.OneArgSink(s3)
//}
//
func toInterface(s core.Source) interface{} {
	return s
}

func dereference(sp *core.Source) core.Source {
	return *sp
}
