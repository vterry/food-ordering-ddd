package vo

type ID struct {
	value string
}

func NewID(value string) ID {
	return ID{value: value}
}

func (id ID) String() string {
	return id.value
}

func (id ID) IsEmpty() bool {
	return id.value == ""
}

func (id ID) Equals(other ID) bool {
	return id.value == other.value
}
