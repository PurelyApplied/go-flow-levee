package types

// Source will be configured to be detected as a source.
type Source struct {
	Data string
	ID   int
}

func (s Source) getID() int {
	return s.ID
}

func (s Source) getData() string {
	return s.Data
}

// Innocuous will _not_ be configured to be a source.
type Innocuous struct {
	Data string
	ID int
}

func (i Innocuous) getID() int {
	return i.ID
}

func (i Innocuous) getData() string {
	return i.Data
}

