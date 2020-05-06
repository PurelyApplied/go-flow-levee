package sources

type Source struct {
	data string
	id   int
}

type Public struct {
	data string
	id   int
}

func testParam(s Source) { // want "source identified: param" "source identified: alloc"
	ptr := &s  // TODO want "source identified: something..."
	_ = ptr
}

func testRefParam(ptr *Source) { // want "source identified: param"
	s := *ptr  // want "source identified: alloc"
	_ = s
}

func testFreeVar() {
	// These trigger twice, once for an Alloc and once as a FreeVar
	var s Source   // want "source identified: freeVar" "source identified: alloc"
	var ps *Source // want "source identified: freeVar" "source identified: alloc"

	go func() {
		var x interface{} = s // TODO want "source identified"
		x = ps                // TODO want "source identified"
		_ = x
	}()

}
