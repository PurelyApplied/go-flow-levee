package foo

type Source struct {
	data string
	id   int
}

type Public struct {
	data string
	id   int
}

func testParam(s Source) { // want "source identified"

}

func testRefParam(s *Source) { // want "source identified"

}

func testFreeVar() {
	// These trigger twice, once for an Alloc and once as a FreeVar
	var s Source   // want "source identified" "source identified"
	var ps *Source // want "source identified" "source identified"

	go func() {
		var x interface{} = s // TODO(?) want "source identified"
		x = ps
		_ = x
	}()

}
