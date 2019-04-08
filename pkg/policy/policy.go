package policy

import (
	"encoding/json"
	"fmt"
	"io"
)

// DefaultVersion - default policy version as per AWS S3 specification.
const DefaultVersion = "2012-10-17"

type Policy struct {
	ID         ID          `json:"ID,omitempty"`
	Version    string      `json:"version"`
	Statements []Statement `json:"Statement"`
}

// IsAllowed - checks given policy args is allowed to continue the Rest API.
func (policy Policy) IsAllowed(args Args) bool {
	// Check all deny statements. If any one statement denies, return false.
	for _, statement := range policy.Statements {
		if statement.Effect == Deny {
			if !statement.IsAllowed(args) {
				return false
			}
		}
	}

	// For owner, its allowed by default.
	if args.IsOwner {
		return true
	}

	// Check all allow statements. If any one statement allows, return true.
	for _, statement := range policy.Statements {
		if statement.Effect == Allow {
			if statement.IsAllowed(args) {
				return true
			}
		}
	}

	return false
}

// IsEmpty - returns whether policy is empty or not.
func (policy Policy) IsEmpty() bool {
	return len(policy.Statements) == 0
}

// isValid - checks if Policy is valid or not.
func (policy Policy) isValid() error {
	if policy.Version != DefaultVersion && policy.Version != "" {
		return fmt.Errorf("invalid version '%v'", policy.Version)
	}

	for _, statement := range policy.Statements {
		if err := statement.isValid(); err != nil {
			return err
		}
	}

	for i := range policy.Statements {
		for _, statement := range policy.Statements[i+1:] {
			principals := policy.Statements[i].Principal.Intersection(statement.Principal)
			if principals.IsEmpty() {
				continue
			}

			actions := policy.Statements[i].Actions.Intersection(statement.Actions)
			if len(actions) == 0 {
				continue
			}

			resources := policy.Statements[i].Resources.Intersection(statement.Resources)
			if len(resources) == 0 {
				continue
			}

			if policy.Statements[i].Conditions.String() != statement.Conditions.String() {
				continue
			}

			return fmt.Errorf("duplicate principal %v, actions %v, resouces %v found in statements %v, %v",
				principals, actions, resources, policy.Statements[i], statement)
		}
	}

	return nil
}

// MarshalJSON - encodes Policy to JSON data.
func (policy Policy) MarshalJSON() ([]byte, error) {
	if err := policy.isValid(); err != nil {
		return nil, err
	}

	// subtype to avoid recursive call to MarshalJSON()
	type subPolicy Policy
	return json.Marshal(subPolicy(policy))
}

func (policy Policy) GetBSON() (interface{}, error) {
	if err := policy.isValid(); err != nil {
		return nil, err
	}

	return struct {
		ID         ID          `json:"ID,omitempty" bson:"id,omitempty"`
		Version    string      `json:"version" bson:"version"`
		Statements []Statement `json:"Statement" bson:"statement"`
	}{
		ID:         policy.ID,
		Version:    policy.Version,
		Statements: policy.Statements,
	}, nil
}

// Validate - validates all statements are for given bucket or not.
func (policy Policy) Validate(bucketName string) error {
	if err := policy.isValid(); err != nil {
		return err
	}

	for _, statement := range policy.Statements {
		if err := statement.Validate(bucketName); err != nil {
			return err
		}
	}

	return nil
}

// ParseConfig - parses data in given reader to Policy.
func ParseConfig(reader io.Reader, bucketName string) (*Policy, error) {
	var policy Policy

	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&policy); err != nil {
		return nil, err
	}

	err := policy.Validate(bucketName)
	return &policy, err
}
