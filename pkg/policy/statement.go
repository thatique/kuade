package policy

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/thatique/kuade/pkg/policy/condition"
)

// Statement - policy statement.
type Statement struct {
	SID        ID                  `json:"Sid,omitempty" bson:"sid,omitempty"`
	Effect     Effect              `json:"Effect" bson:"effect"`
	Principal  Principal           `json:"Principal" bson:"principal"`
	Actions    ActionSet           `json:"Action" bson:"action"`
	Resources  ResourceSet         `json:"Resource" bson:"resource"`
	Conditions condition.Functions `json:"Condition,omitempty" bson:"condition,omitempty"`
}

// IsAllowed - checks given policy args is allowed to continue the Rest API.
func (statement Statement) IsAllowed(args Args) bool {
	check := func() bool {
		if !statement.Principal.Match(args.AccountName) {
			return false
		}

		if !statement.Actions.Contains(args.Action) {
			return false
		}

		resource := args.ResourceName
		if args.ObjectName != "" {
			if !strings.HasPrefix(args.ObjectName, "/") {
				resource += "/"
			}

			resource += args.ObjectName
		}

		if !statement.Resources.Match(resource, args.ConditionValues) {
			return false
		}

		return statement.Conditions.Evaluate(args.ConditionValues)
	}

	return statement.Effect.IsAllowed(check())
}

// isValid - checks whether statement is valid or not.
func (statement Statement) isValid() error {
	if !statement.Effect.IsValid() {
		return fmt.Errorf("invalid Effect %v", statement.Effect)
	}

	if !statement.Principal.IsValid() {
		return fmt.Errorf("invalid Principal %v", statement.Principal)
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
func (statement Statement) Validate(bucketName string) error {
	if err := statement.isValid(); err != nil {
		return err
	}

	return statement.Resources.Validate(bucketName)
}

// NewStatement - creates new statement.
func NewStatement(effect Effect, principal Principal, actionSet ActionSet, resourceSet ResourceSet, conditions condition.Functions) Statement {
	return Statement{
		Effect:     effect,
		Principal:  principal,
		Actions:    actionSet,
		Resources:  resourceSet,
		Conditions: conditions,
	}
}
