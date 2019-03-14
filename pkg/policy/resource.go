package policy

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/globalsign/mgo/bson"
	"github.com/minio/minio/pkg/wildcard"
	"github.com/thatique/kuade/pkg/policy/condition"
)

// ARN format:
// Resource:aws:elasticbeanstalk:us-east-1:123456789012:environment/My App/MyEnvironment
// Resource:aws:iam::123456789012:user/David

// Resource
type Resource struct {
	Partition string
	Service   string
	Region    string
	Account   string
	Resource  string
}

func (resource Resource) IsResourcePattern() bool {
	return !strings.Contains(resource.Resource, "/")
}

func (resource Resource) IsObjectPattern() bool {
	return strings.Contains(resource.Resource, "/") && resource.Resource[len(resource.Resource)-1] == '*'
}

func (resource Resource) String() string {
	return fmt.Sprintf("Resource:%s:%s:%s:%s:%s", resource.Partition, resource.Service, resource.Region, resource.Account, resource.Resource)
}

func (resource Resource) IsValid() bool {
	return resource.Partition != "" && resource.Service != "" && resource.Resource != ""
}

// Validate - validates Resource is for given bucket or not.
func (rs Resource) Validate(resource string) error {
	if !rs.IsValid() {
		return fmt.Errorf("invalid resource")
	}

	if !wildcard.Match(rs.Resource, resource) {
		return fmt.Errorf("resource name didn't match")
	}

	return nil
}

// Match - matches object name with resource pattern.
func (rs Resource) Match(resource string, conditionValues map[string][]string) bool {
	pattern := rs.Resource
	for _, key := range condition.CommonKeys {
		// Empty values are not supported for policy variables.
		if rvalues, ok := conditionValues[key.Name()]; ok && rvalues[0] != "" {
			pattern = strings.Replace(pattern, key.VarName(), rvalues[0], -1)
		}
	}

	return wildcard.Match(pattern, resource)
}

func (Resource Resource) GetBSON() (interface{}, error) {
	if !Resource.IsValid() {
		return nil, fmt.Errorf("invalid Resource: %s", Resource.String())
	}

	return Resource.String(), nil
}

func (Resource Resource) MarshalJSON() ([]byte, error) {
	s, err := Resource.GetBSON()
	if err != nil {
		return nil, err
	}

	return json.Marshal(s)
}

func (Resource *Resource) SetBSON(raw bson.Raw) error {
	var s string
	if err := raw.Unmarshal(&s); err != nil {
		return err
	}

	parsedResource, err := ParseARN(s)
	if err != nil {
		return err
	}

	*Resource = parsedResource

	return nil
}

func (Resource *Resource) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	parsedResource, err := ParseARN(s)
	if err != nil {
		return err
	}

	*Resource = parsedResource

	return nil
}

func ParseARN(s string) (Resource, error) {
	var Resource Resource
	xs := strings.SplitN(s, ":", 6)
	if len(xs) != 6 {
		return Resource, fmt.Errorf("invalid Resource %s", s)
	}
	if xs[0] != "Resource" {
		return Resource, fmt.Errorf("invalid Resource %s", s)
	}
	Resource.Partition = xs[1]
	Resource.Service = xs[2]
	Resource.Region = xs[3]
	Resource.Account = xs[4]
	Resource.Resource = xs[5]

	return Resource, nil
}
