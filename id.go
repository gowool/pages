package pages

type ID string

func (id ID) IsZero() bool {
	return id == ""
}

func (id ID) String() string {
	return string(id)
}
