package arguments

import (
	"example.com/types"
)


func testSourceFromParamByReference(s *types.Source) {
	types.Sink("Source in the parameter %v", s) // want "a source has reached a sink"
}

func testSourceMethodFromParamByReference(s *types.Source) {
	types.Sink("Source in the parameter %v", s.Data) // want "a source has reached a sink"
}

func testSourceFromParamByReferenceInfo(s *types.Source) {
	types.Sink(s) // want "a source has reached a sink"
}

func testSourceFromParamByValue(s types.Source) {
	types.Sink("Source in the parameter %v", s) // want "a source has reached a sink"
}

func testUpdatedSource(s *types.Source) {
	s.Data = "updated"
	types.Sink("Updated %v", s) // want "a source has reached a sink"
}

func testSourceFromAPointerCopy(s *types.Source) {
	cp := s
	types.Sink("Pointer copy of the source %v", cp) // want "a source has reached a sink"
}
