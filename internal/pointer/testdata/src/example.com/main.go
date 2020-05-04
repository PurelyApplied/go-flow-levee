package main


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



func directCall(s Source) { // want "identified as source"
	Sink(s)
}

func interfaceCall(s interface{}) {
	Sink(s)
}

func main() {
	s := Source{SourceData: "secret"}  // want "identified as source"
	directCall(s)
	interfaceCall(s)
}
