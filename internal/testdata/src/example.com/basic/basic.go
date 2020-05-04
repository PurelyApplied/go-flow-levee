package basic

import "io"

type Source struct {
	data string
	id   int
}

func Sink(args ...interface{}) {}

func Sinkf(format string, args ...interface{}) {}

func FSinkf(writer io.Writer, args ...interface{}) {}

func SingleArgSink(interface{}) {}

func TestSinks(s Source, writer io.Writer) {
	Sink(s)                  // want "a source has reached a sink"
	Sinkf("a source: %v", s) // want "a source has reached a sink"
	FSinkf(writer, s)        // want "a source has reached a sink"
	SingleArgSink(s)         // TODO want "a source has reached a sink"

	Sink([]interface{}{s, s, s}...) // TODO want "a source has reached a sink"
	Sink([]interface{}{s, s, s})    // TODO want "a source has reached a sink"
}

func TestSinksWithRef(s *Source, writer io.Writer) {
	Sink(s)                  // want "a source has reached a sink"
	Sinkf("a source: %v", s) // want "a source has reached a sink"
	FSinkf(writer, s)        // want "a source has reached a sink"
	SingleArgSink(s)         // TODO want "a source has reached a sink"

	Sink([]interface{}{s, s, s}...) // TODO want "a source has reached a sink"
	Sink([]interface{}{s, s, s})    // TODO want "a source has reached a sink"
}
