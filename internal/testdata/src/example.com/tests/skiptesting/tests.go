package skiptesting

import (
	"io"
	"testing"

	"example.com/core"
)

// Packages importing testing are not examined.  No reports should be emitted.
func callSinks(s core.Source, writer io.Writer) {
	core.Sink(s)
	core.Sinkf("a source: %v", s)
	core.FSinkf(writer, s)
}

func TestSink(t *testing.T) {
	callSinks(core.Source{}, nil)
}
