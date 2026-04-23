package vo

import (
	"encoding/json"
)

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

func (id ID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.value)
}

func (id *ID) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	id.value = s
	return nil
}
