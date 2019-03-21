package policy

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/globalsign/mgo/bson"
	"github.com/minio/minio-go/pkg/set"
	"github.com/thatique/kuade/pkg/arn"
)

// ResourceSet - set of ARN in policy statement.
type ResourceSet map[arn.ARN]struct{}

// bucketResourceExists - checks if at least one resource exists in the set.
func (resourceSet ResourceSet) resourceExists() bool {
	for resource := range resourceSet {
		if resource.IsResourcePattern() {
			return true
		}
	}

	return false
}

// objectResourceExists - checks if at least one object resource exists in the set.
func (resourceSet ResourceSet) objectResourceExists() bool {
	for resource := range resourceSet {
		if resource.IsObjectPattern() {
			return true
		}
	}

	return false
}

// Add - adds resource to resource set.
func (resourceSet ResourceSet) Add(resource arn.ARN) {
	resourceSet[resource] = struct{}{}
}

// Intersection - returns resouces available in both ResourcsSet.
func (resourceSet ResourceSet) Intersection(sset ResourceSet) ResourceSet {
	nset := NewResourceSet()
	for k := range resourceSet {
		if _, ok := sset[k]; ok {
			nset.Add(k)
		}
	}

	return nset
}

// MarshalJSON - encodes ResourceSet to JSON data.
func (resourceSet ResourceSet) MarshalJSON() ([]byte, error) {
	if len(resourceSet) == 0 {
		return nil, fmt.Errorf("empty resource set")
	}

	resources := []arn.ARN{}
	for resource := range resourceSet {
		resources = append(resources, resource)
	}

	return json.Marshal(resources)
}

func (resourceSet ResourceSet) GetBSON() (interface{}, error) {
	if len(resourceSet) == 0 {
		return nil, fmt.Errorf("empty resource set")
	}

	resources := []arn.ARN{}
	for resource := range resourceSet {
		resources = append(resources, resource)
	}

	return resources, nil
}

// Match - matches object name with anyone of resource pattern in resource set.
func (resourceSet ResourceSet) Match(resource string, conditionValues map[string][]string) bool {
	for r := range resourceSet {
		if r.Match(resource, conditionValues) {
			return true
		}
	}

	return false
}

func (resourceSet ResourceSet) String() string {
	resources := []string{}
	for resource := range resourceSet {
		resources = append(resources, resource.String())
	}
	sort.Strings(resources)

	return fmt.Sprintf("%v", resources)
}

// UnmarshalJSON - decodes JSON data to ResourceSet.
func (resourceSet *ResourceSet) UnmarshalJSON(data []byte) error {
	var sset set.StringSet
	if err := json.Unmarshal(data, &sset); err != nil {
		return err
	}

	*resourceSet = make(ResourceSet)
	for _, s := range sset.ToSlice() {
		resource, err := arn.ParseARN(s)
		if err != nil {
			return err
		}

		if _, found := (*resourceSet)[resource]; found {
			return fmt.Errorf("duplicate resource '%v' found", s)
		}

		resourceSet.Add(resource)
	}

	return nil
}

func (resourceSet *ResourceSet) SetBSON(raw bson.Raw) error {
	var sset set.StringSet
	if err := raw.Unmarshal(&sset); err != nil {
		return err
	}

	*resourceSet = make(ResourceSet)
	for _, s := range sset.ToSlice() {
		resource, err := arn.ParseARN(s)
		if err != nil {
			return err
		}

		if _, found := (*resourceSet)[resource]; found {
			return fmt.Errorf("duplicate resource '%v' found", s)
		}

		resourceSet.Add(resource)
	}

	return nil
}

// Validate - validates ResourceSet is for given bucket or not.
func (resourceSet ResourceSet) Validate(bucketName string) error {
	for resource := range resourceSet {
		if err := resource.Validate(bucketName); err != nil {
			return err
		}
	}

	return nil
}

// NewResourceSet - creates new resource set.
func NewResourceSet(resources ...arn.ARN) ResourceSet {
	resourceSet := make(ResourceSet)
	for _, resource := range resources {
		resourceSet.Add(resource)
	}

	return resourceSet
}
