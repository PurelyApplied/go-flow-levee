package foo

type Source struct {
	data string
	id   int
}

type Public struct {
	data string
	id   int
}

func testParam(s Source) { // want "source identified: alloc"

}

func testRefParam(s *Source) { // want "source identified: param"

}

func testFreeVar() {
	// These trigger twice, once for an Alloc and once as a FreeVar
	var s Source   // want "source identified: freeVar" "source identified: alloc"
	var ps *Source // want "source identified: freeVar" "source identified: alloc"

	go func() {
		var x interface{} = s // TODO(?) want "source identified"
		x = ps
		_ = x
	}()

}
