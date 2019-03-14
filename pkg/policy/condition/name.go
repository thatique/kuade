package condition

import (
	"encoding/json"
	"fmt"

	"github.com/globalsign/mgo/bson"
)

type name string

const (
	stringEquals              name = "StringEquals"
	stringNotEquals                = "StringNotEquals"
	stringEqualsIgnoreCase         = "StringEqualsIgnoreCase"
	stringNotEqualsIgnoreCase      = "StringNotEqualsIgnoreCase"
	stringLike                     = "StringLike"
	stringNotLike                  = "StringNotLike"
	binaryEquals                   = "BinaryEquals"
	ipAddress                      = "IpAddress"
	notIPAddress                   = "NotIpAddress"
	null                           = "Null"
	boolean                        = "Bool"
)

var supportedConditions = []name{
	stringEquals,
	stringNotEquals,
	stringEqualsIgnoreCase,
	stringNotEqualsIgnoreCase,
	binaryEquals,
	stringLike,
	stringNotLike,
	ipAddress,
	notIPAddress,
	null,
	boolean,
	// Add new conditions here.
}

// IsValid - checks if name is valid or not
func (n name) IsValid() bool {
	for _, supn := range supportedConditions {
		if n == supn {
			return true
		}
	}

	return false
}

func (n name) MarshalJSON() ([]byte, error) {
	if !n.IsValid() {
		return nil, fmt.Errorf("invalid condition name '%v'", n)
	}

	return json.Marshal(string(n))
}

func (n *name) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	parsedName, err := parseName(s)
	if err != nil {
		return err
	}

	*n = parsedName
	return nil
}

func (n name) GetBSON() (interface{}, error) {
	if !n.IsValid() {
		return nil, fmt.Errorf("invalid condition name '%v'", n)
	}

	return string(n), nil
}

func (n *name) SetBSON(raw bson.Raw) error {
	var s string
	if err := raw.Unmarshal(&s); err != nil {
		return err
	}

	parsedName, err := parseName(s)
	if err != nil {
		return err
	}

	*n = parsedName
	return nil
}

func parseName(s string) (name, error) {
	n := name(s)

	if n.IsValid() {
		return n, nil
	}

	return n, fmt.Errorf("invalid condition name '%v'", s)
}
