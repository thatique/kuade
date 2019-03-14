package policy

import (
	"encoding/json"
	"fmt"

	"github.com/globalsign/mgo/bson"
	"github.com/minio/minio-go/pkg/set"
	"github.com/minio/minio/pkg/wildcard"
)

// Principal - policy principal.
type Principal struct {
	AWS set.StringSet
}

// IsValid - checks whether Principal is valid or not.
func (p Principal) IsValid() bool {
	return len(p.AWS) != 0
}

// Intersection - returns principals available in both Principal.
func (p Principal) Intersection(principal Principal) set.StringSet {
	return p.AWS.Intersection(principal.AWS)
}

// MarshalJSON - encodes Principal to JSON data.
func (p Principal) MarshalJSON() ([]byte, error) {
	if !p.IsValid() {
		return nil, fmt.Errorf("invalid principal %v", p)
	}

	// subtype to avoid recursive call to MarshalJSON()
	type subPrincipal Principal
	sp := subPrincipal(p)
	return json.Marshal(sp)
}

//
func (p Principal) GetBSON() (interface{}, error) {
	if !p.IsValid() {
		return nil, fmt.Errorf("invalid principal %v", p)
	}

	return struct {
		AWS set.StringSet
	}{
		AWS: p.AWS,
	}, nil
}

// Match - matches given principal is wildcard matching with Principal.
func (p Principal) Match(principal string) bool {
	for _, pattern := range p.AWS.ToSlice() {
		if wildcard.MatchSimple(pattern, principal) {
			return true
		}
	}

	return false
}

// UnmarshalJSON - decodes JSON data to Principal.
func (p *Principal) UnmarshalJSON(data []byte) error {
	// subtype to avoid recursive call to UnmarshalJSON()
	type subPrincipal Principal
	var sp subPrincipal

	if err := json.Unmarshal(data, &sp); err != nil {
		var s string
		if err = json.Unmarshal(data, &s); err != nil {
			return err
		}

		if s != "*" {
			return fmt.Errorf("invalid principal '%v'", s)
		}

		sp.AWS = set.CreateStringSet("*")
	}

	*p = Principal(sp)

	return nil
}

func (p *Principal) SetBSON(raw bson.Raw) error {
	type subPrincipal Principal
	var sp subPrincipal

	if err := raw.Unmarshal(&sp); err != nil {
		var s string
		if err = raw.Unmarshal(&s); err != nil {
			return err
		}

		if s != "*" {
			return fmt.Errorf("invalid principal '%v'", s)
		}

		sp.AWS = set.CreateStringSet("*")
	}

	*p = Principal(sp)
	return nil
}

// NewPrincipal - creates new Principal.
func NewPrincipal(principals ...string) Principal {
	return Principal{AWS: set.CreateStringSet(principals...)}
}
