package foo

type Source struct {
	SourceData string
	OtherData  int
}

var globalSource Source // TODO want "identified as source"

func bar() {
	var s Source // want "identified as source"
	_ = s
}

func Sink(interface{}){}