package arguments

import (
	"example.com/core"
)

func testSourceFromParamByReference(s *core.Source) {
	core.Sink("Source in the parameter %v", s) // want "a source has reached a sink"
}

func testSourceMethodFromParamByReference(s *core.Source) {
	core.Sink("Source in the parameter %v", s.Data) // want "a source has reached a sink"
}

func testSourceFromParamByReferenceInfo(s *core.Source) {
	core.Sink(s) // want "a source has reached a sink"
}

func testSourceFromParamByValue(s core.Source) {
	core.Sink("Source in the parameter %v", s) // want "a source has reached a sink"
}

func testUpdatedSource(s *core.Source) {
	s.Data = "updated"
	core.Sink("Updated %v", s) // want "a source has reached a sink"
}

func testSourceFromAPointerCopy(s *core.Source) {
	cp := s
	core.Sink("Pointer copy of the source %v", cp) // want "a source has reached a sink"
}
