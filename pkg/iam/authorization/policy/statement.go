package iampolicy

import (
	"encoding/json"
	"fmt"

	"github.com/thatique/kuade/pkg/iam/authorization/authorizer"
	"github.com/thatique/kuade/pkg/policy"
	"github.com/thatique/kuade/pkg/policy/condition"
)

// Statement - policy statement
type Statement struct {
	SID        policy.ID           `json:"SID,omitempty"`
	Effect     policy.Effect       `json:"Effect"`
	Actions    ActionSet           `json:"Actions"`
	Resources  policy.ResourceSet  `json:"Resources"`
	Conditions condition.Functions `json:"Conditions,omitempty"`
}

// IsAllowed - Check if the given args alllowed
func (statement Statement) IsAllowed(args authorizer.Args) bool {
	check := func() bool {
		if !statement.Actions.Match(Action(args.Action)) {
			return false
		}
		if !statement.Resources.Match(args.ARN.Resource, args.Metadata) {
			return false
		}

		return statement.Conditions.Evaluate(args.Metadata)
	}

	return statement.Effect.IsAllowed(check())
}

// isValid - checks whether statement is valid or not.
func (statement Statement) isValid() error {
	if !statement.Effect.IsValid() {
		return fmt.Errorf("invalid Effect %v", statement.Effect)
	}

	if len(statement.Actions) == 0 {
		return fmt.Errorf("Action must not be empty")
	}

	if len(statement.Resources) == 0 {
		return fmt.Errorf("Resource must not be empty")
	}

	return nil
}

// MarshalJSON - encodes JSON data to Statement.
func (statement Statement) MarshalJSON() ([]byte, error) {
	if err := statement.isValid(); err != nil {
		return nil, err
	}

	// subtype to avoid recursive call to MarshalJSON()
	type subStatement Statement
	ss := subStatement(statement)
	return json.Marshal(ss)
}

// UnmarshalJSON - decodes JSON data to Statement.
func (statement *Statement) UnmarshalJSON(data []byte) error {
	// subtype to avoid recursive call to UnmarshalJSON()
	type subStatement Statement
	var ss subStatement

	if err := json.Unmarshal(data, &ss); err != nil {
		return err
	}

	s := Statement(ss)
	if err := s.isValid(); err != nil {
		return err
	}

	*statement = s

	return nil
}

// Validate - validates Statement is for given bucket or not.
func (statement Statement) Validate() error {
	return statement.isValid()
}

// NewStatement - creates new statement.
func NewStatement(effect policy.Effect, actionSet ActionSet, resourceSet policy.ResourceSet, conditions condition.Functions) Statement {
	return Statement{
		Effect:     effect,
		Actions:    actionSet,
		Resources:  resourceSet,
		Conditions: conditions,
	}
}
