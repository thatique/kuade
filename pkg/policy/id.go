package policy

import (
	"encoding/json"
	"fmt"
	"unicode/utf8"
)

// ID - policy ID.
type ID string

// IsValid - checks if ID is valid or not.
func (id ID) IsValid() bool {
	return utf8.ValidString(string(id))
}

// MarshalJSON - encodes ID to JSON data.
func (id ID) MarshalJSON() ([]byte, error) {
	if !id.IsValid() {
		return nil, fmt.Errorf("invalid ID %v", id)
	}

	return json.Marshal(string(id))
}

// UnmarshalJSON - decodes JSON data to ID.
func (id *ID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	i := ID(s)
	if !i.IsValid() {
		return fmt.Errorf("invalid ID %v", s)
	}

	*id = i

	return nil
}
