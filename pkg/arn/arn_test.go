package arn

import (
	"testing"
)

func TestParseARN(t *testing.T) {
	xs := []string{
		"arn:aws:elasticbeanstalk:us-east-1:123456789012:environment/My App/MyEnvironment",
		"arn:aws:iam::123456789012:user/David",
	}
	for i, str := range xs {
		arn, err := ParseARN(str)
		if err != nil {
			t.Fatalf("expected valid arn at %d: %v", i, err)
		}
		if arn.String() != str {
			t.Fatalf("parsed: %s arn and string: %s not equal.", arn.String(), str)
		}
	}
}
