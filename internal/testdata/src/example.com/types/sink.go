package types

import "io"

func Sink(args ...interface{}) {}

func Sinkf(format string, args ...interface{}) {}

func FSinkf(writer io.Writer, args ...interface{}) {}

func SingleArgSink(interface{}) {}


