package whitelisted

import (
	"io"

	"example.com/core"
)

// This file is whitelisted in analysis configuration.  No reports should be emitted.
func TestSinks(s core.Source, writer io.Writer) {
	core.Sink(s)
	core.Sinkf("a source: %v", s)
	core.FSinkf(writer, s)
}
