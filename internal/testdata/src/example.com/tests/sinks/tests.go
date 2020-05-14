package sinks

import (
	"io"

	"example.com/core"
)

func TestSinks(s core.Source, writer io.Writer) {
	var slice []interface{}
	slice = make([]interface{}, 1)
	slice[0] = s
	core.Sink(slice) // want "a source has reached a sink"
}
