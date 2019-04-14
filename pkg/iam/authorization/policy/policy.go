package iampolicy

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/thatique/kuade/pkg/iam/authorization/authorizer"
	"github.com/thatique/kuade/pkg/policy"
)

// DefaultVersion - default policy version as per AWS S3 specification.
const DefaultVersion = "2012-10-17"

var _ authorizer.Authorizer = Policy{}

// Policy authorizer
type Policy struct {
	ID         policy.ID   `json:"ID,omitempty"`
	Version    string      `json:"Version"`
	Statements []Statement `json:"Statement"`
}

// Authorize implements `authorizer.Authorizer` interface. Policy never return
// `DecisionNoOpinion`
func (iamp Policy) Authorize(args authorizer.Args) (authorized authorizer.Decision, reason string, err error) {
	if iamp.IsAllowed(args) {
		return authorizer.DecisionAllow, "", nil
	}
	return authorizer.DecisionDeny, "", nil
}

// IsAllowed evaluate policy for the given args
func (iamp Policy) IsAllowed(args authorizer.Args) bool {
	for _, statement := range iamp.Statements {
		if statement.Effect == policy.Deny {
			if !statement.IsAllowed(args) {
				return false
			}
		}
	}

	// first check if this arg is owner
	if args.IsOwner {
		return true
	}

	for _, statement := range iamp.Statements {
		if statement.Effect == policy.Allow {
			if statement.IsAllowed(args) {
				return true
			}
		}
	}

	return false
}

// IsEmpty - returns whether policy is empty or not.
func (iamp Policy) IsEmpty() bool {
	return len(iamp.Statements) == 0
}

// isValid - checks if Policy is valid or not.
func (iamp Policy) isValid() error {
	if iamp.Version != DefaultVersion && iamp.Version != "" {
		return fmt.Errorf("invalid version '%v'", iamp.Version)
	}

	for _, statement := range iamp.Statements {
		if err := statement.isValid(); err != nil {
			return err
		}
	}

	for i := range iamp.Statements {
		for _, statement := range iamp.Statements[i+1:] {
			actions := iamp.Statements[i].Actions.Intersection(statement.Actions)
			if len(actions) == 0 {
				continue
			}

			resources := iamp.Statements[i].Resources.Intersection(statement.Resources)
			if len(resources) == 0 {
				continue
			}

			if iamp.Statements[i].Conditions.String() != statement.Conditions.String() {
				continue
			}

			return fmt.Errorf("duplicate actions %v, resources %v found in statements %v, %v",
				actions, resources, iamp.Statements[i], statement)
		}
	}

	return nil
}

// MarshalJSON - encodes Policy to JSON data.
func (iamp Policy) MarshalJSON() ([]byte, error) {
	if err := iamp.isValid(); err != nil {
		return nil, err
	}

	// subtype to avoid recursive call to MarshalJSON()
	type subPolicy Policy
	return json.Marshal(subPolicy(iamp))
}

// UnmarshalJSON - decodes JSON data to Iamp.
func (iamp *Policy) UnmarshalJSON(data []byte) error {
	// subtype to avoid recursive call to UnmarshalJSON()
	type subPolicy Policy
	var sp subPolicy
	if err := json.Unmarshal(data, &sp); err != nil {
		return err
	}

	p := Policy(sp)
	if err := p.isValid(); err != nil {
		return err
	}

	*iamp = p

	return nil
}

// Validate - validates all statements are for given bucket or not.
func (iamp Policy) Validate() error {
	if err := iamp.isValid(); err != nil {
		return err
	}

	for _, statement := range iamp.Statements {
		if err := statement.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ParseConfig - parses data in given reader to Iamp.
func ParseConfig(reader io.Reader) (*Policy, error) {
	var iamp Policy

	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&iamp); err != nil {
		return nil, err
	}

	return &iamp, iamp.Validate()
}
