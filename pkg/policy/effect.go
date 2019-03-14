package policy

import (
	"encoding/json"
	"fmt"

	"github.com/globalsign/mgo/bson"
)

type Effect string

const (
	// Allow - allow effect.
	Allow Effect = "Allow"

	// Deny - deny effect.
	Deny = "Deny"
)

// IsAllowed - returns if given check is allowed or not.
func (effect Effect) IsAllowed(b bool) bool {
	if effect == Allow {
		return b
	}

	return !b
}

// IsValid - checks if Effect is valid or not
func (effect Effect) IsValid() bool {
	switch effect {
	case Allow, Deny:
		return true
	}

	return false
}

// MarshalJSON - encodes Effect to JSON data.
func (effect Effect) MarshalJSON() ([]byte, error) {
	if !effect.IsValid() {
		return nil, fmt.Errorf("invalid effect '%v'", effect)
	}

	return json.Marshal(string(effect))
}

func (effect Effect) GetBSON() (interface{}, error) {
	if !effect.IsValid() {
		return nil, fmt.Errorf("invalid effect keu %v", effect)
	}

	return string(effect), nil
}

func (effect *Effect) SetBSON(raw bson.Raw) error {
	var s string
	if err := raw.Unmarshal(&s); err != nil {
		return err
	}

	e := Effect(s)
	if !e.IsValid() {
		return fmt.Errorf("invalid effect '%v'", s)
	}

	*effect = e

	return nil
}

// UnmarshalJSON - decodes JSON data to Effect.
func (effect *Effect) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	e := Effect(s)
	if !e.IsValid() {
		return fmt.Errorf("invalid effect '%v'", s)
	}

	*effect = e

	return nil
}
