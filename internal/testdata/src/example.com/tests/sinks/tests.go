package sinks

import (
	"io"

	"example.com/types"
)

func TestSinks(s types.Source, writer io.Writer) {
	types.Sink(s)                  // want "a source has reached a sink"
	types.Sinkf("a source: %v", s) // want "a source has reached a sink"
	types.FSinkf(writer, s)        // want "a source has reached a sink"
	types.SingleArgSink(s)         // TODO want "a source has reached a sink"

	types.Sink([]interface{}{s, s, s}...) // TODO want "a source has reached a sink"
	types.Sink([]interface{}{s, s, s})    // TODO want "a source has reached a sink"
}

func TestSinksWithRef(s *types.Source, writer io.Writer) {
	types.Sink(s)                  // want "a source has reached a sink"
	types.Sinkf("a source: %v", s) // want "a source has reached a sink"
	types.FSinkf(writer, s)        // want "a source has reached a sink"
	types.SingleArgSink(s)         // TODO want "a source has reached a sink"

	types.Sink([]interface{}{s, s, s}...) // TODO want "a source has reached a sink"
	types.Sink([]interface{}{s, s, s})    // TODO want "a source has reached a sink"
}

func TestSinksInnocuous(innoc types.Innocuous, writer io.Writer) {
	types.Sink(innoc)
	types.Sinkf("a source: %v", innoc)
	types.FSinkf(writer, innoc)
	types.SingleArgSink(innoc)

	types.Sink([]interface{}{innoc, innoc, innoc}...)
	types.Sink([]interface{}{innoc, innoc, innoc})
}

func TestSinksWithInnocuousRef(innoc *types.Innocuous, writer io.Writer) {
	types.Sink(innoc)
	types.Sinkf("a source: %v", innoc)
	types.FSinkf(writer, innoc)
	types.SingleArgSink(innoc)

	types.Sink([]interface{}{innoc, innoc, innoc}...)
	types.Sink([]interface{}{innoc, innoc, innoc})
}
