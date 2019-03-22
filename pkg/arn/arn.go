package arn

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/globalsign/mgo/bson"
	"github.com/minio/minio/pkg/wildcard"
	"github.com/thatique/kuade/pkg/policy/condition"
)

type ARN struct {
	Partition string
	Service   string
	Region    string
	Account   string
	//
	Resource string
}

func (arn ARN) IsResourcePattern() bool {
	return !strings.Contains(arn.Resource, "/")
}

func (arn ARN) IsObjectPattern() bool {
	return strings.Contains(arn.Resource, "/") && arn.Resource[len(arn.Resource)-1] == '*'
}

func (arn ARN) String() string {
	return fmt.Sprintf("arn:%s:%s:%s:%s:%s", arn.Partition, arn.Service, arn.Region, arn.Account, arn.Resource)
}

func (arn ARN) IsValid() bool {
	return arn.Partition != "" && arn.Service != "" && arn.Resource != ""
}

func (arn ARN) IsOwnedBy(account string) bool {
	return arn.Account == account && arn.Account != ""
}

func (arn ARN) Validate(resource string) error {
	if !arn.IsValid() {
		return fmt.Errorf("invalid resource")
	}

	if !wildcard.Match(arn.Resource, resource) {
		return fmt.Errorf("resource name didn't match")
	}

	return nil
}

// Match - matches object name with resource pattern.
func (arn ARN) Match(resource string, conditionValues map[string][]string) bool {
	pattern := arn.Resource
	for _, key := range condition.CommonKeys {
		// Empty values are not supported for policy variables.
		if rvalues, ok := conditionValues[key.Name()]; ok && rvalues[0] != "" {
			pattern = strings.Replace(pattern, key.VarName(), rvalues[0], -1)
		}
	}

	return wildcard.Match(pattern, resource)
}

func (arn ARN) GetBSON() (interface{}, error) {
	if !arn.IsValid() {
		return nil, fmt.Errorf("invalid ARN: %s", arn.String())
	}

	return arn.String(), nil
}

func (arn ARN) MarshalJSON() ([]byte, error) {
	s, err := arn.GetBSON()
	if err != nil {
		return nil, err
	}

	return json.Marshal(s)
}

func (arn *ARN) SetBSON(raw bson.Raw) error {
	var s string
	if err := raw.Unmarshal(&s); err != nil {
		return err
	}

	parsedARN, err := ParseARN(s)
	if err != nil {
		return err
	}

	*arn = parsedARN

	return nil
}

func (arn *ARN) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	parsedARN, err := ParseARN(s)
	if err != nil {
		return err
	}

	*arn = parsedARN

	return nil
}

func ParseARN(s string) (ARN, error) {
	var arn ARN
	xs := strings.SplitN(s, ":", 6)
	if len(xs) != 6 {
		return arn, fmt.Errorf("invalid ARN %s", s)
	}
	if xs[0] != "arn" {
		return arn, fmt.Errorf("invalid ARN %s", s)
	}
	arn.Partition = xs[1]
	arn.Service = xs[2]
	arn.Region = xs[3]
	arn.Account = xs[4]
	arn.Resource = xs[5]

	return arn, nil
}
