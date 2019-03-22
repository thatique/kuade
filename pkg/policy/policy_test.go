package policy

import (
	"net"
	"testing"

	"github.com/thatique/kuade/pkg/arn"
	"github.com/thatique/kuade/pkg/policy/condition"
)

func TestIsPolicyAllowed(t *testing.T) {
	case1Policy := Policy{
		Version: DefaultVersion,
		Statements: []Statement{
			NewStatement(
				Allow,
				NewPrincipal("*"),
				NewActionSet(Action("s3:GetBucketLocation"), Action("s3:PutObject")),
				NewResourceSet(arn.ARN{Partition: "aws", Service: "s3", Resource: "*"}),
				condition.NewFunctions(),
			),
		},
	}

	case2Policy := Policy{
		Version: DefaultVersion,
		Statements: []Statement{
			NewStatement(
				Allow,
				NewPrincipal("*"),
				NewActionSet(Action("s3:GetObject"), Action("s3:PutObject")),
				NewResourceSet(arn.ARN{Partition: "aws", Service: "s3", Resource: "mybucket/myobject*"}),
				condition.NewFunctions(),
			)},
	}

	_, IPNet, err := net.ParseCIDR("192.168.1.0/24")
	if err != nil {
		t.Fatalf("unexpected error. %v\n", err)
	}
	func1, err := condition.NewIPAddressFunc(
		condition.AWSSourceIP,
		IPNet,
	)
	if err != nil {
		t.Fatalf("unexpected error. %v\n", err)
	}

	case3Policy := Policy{
		Version: DefaultVersion,
		Statements: []Statement{
			NewStatement(
				Allow,
				NewPrincipal("*"),
				NewActionSet(Action("s3:GetObject"), Action("s3:PutObject")),
				NewResourceSet(arn.ARN{Partition: "aws", Service: "s3", Resource: "mybucket/myobject*"}),
				condition.NewFunctions(func1),
			)},
	}

	case4Policy := Policy{
		Version: DefaultVersion,
		Statements: []Statement{
			NewStatement(
				Deny,
				NewPrincipal("*"),
				NewActionSet(Action("s3:GetObject"), Action("s3:PutObject")),
				NewResourceSet(arn.ARN{Partition: "aws", Service: "s3", Resource: "mybucket/myobject*"}),
				condition.NewFunctions(func1),
			)},
	}

	anonGetBucketLocationArgs := Args{
		AccountName:     "Q3AM3UQ867SPQQA43P2F",
		Action:          Action("s3:GetBucketLocation"),
		ResourceName:    "mybucket",
		ConditionValues: map[string][]string{},
	}

	anonPutObjectActionArgs := Args{
		AccountName:  "Q3AM3UQ867SPQQA43P2F",
		Action:       Action("s3:PutObject"),
		ResourceName: "mybucket",
		ConditionValues: map[string][]string{
			"x-amz-copy-source": {"mybucket/myobject"},
			"SourceIp":          {"192.168.1.10"},
		},
		ObjectName: "myobject",
	}

	anonGetObjectActionArgs := Args{
		AccountName:     "Q3AM3UQ867SPQQA43P2F",
		Action:          Action("s3:GetObject"),
		ResourceName:    "mybucket",
		ConditionValues: map[string][]string{},
		ObjectName:      "myobject",
	}

	getBucketLocationArgs := Args{
		AccountName:     "Q3AM3UQ867SPQQA43P2F",
		Action:          Action("s3:GetBucketLocation"),
		ResourceName:    "mybucket",
		ConditionValues: map[string][]string{},
		IsOwner:         true,
	}

	putObjectActionArgs := Args{
		AccountName:  "Q3AM3UQ867SPQQA43P2F",
		Action:       Action("s3:PutObject"),
		ResourceName: "mybucket",
		ConditionValues: map[string][]string{
			"x-amz-copy-source": {"mybucket/myobject"},
			"SourceIp":          {"192.168.1.10"},
		},
		IsOwner:    true,
		ObjectName: "myobject",
	}

	getObjectActionArgs := Args{
		AccountName:     "Q3AM3UQ867SPQQA43P2F",
		Action:          Action("s3:GetObject"),
		ResourceName:    "mybucket",
		ConditionValues: map[string][]string{},
		IsOwner:         true,
		ObjectName:      "myobject",
	}

	testCases := []struct {
		policy         Policy
		args           Args
		expectedResult bool
	}{
		{case1Policy, anonGetBucketLocationArgs, true},
		{case1Policy, anonPutObjectActionArgs, true},
		{case1Policy, anonGetObjectActionArgs, false},
		{case1Policy, getBucketLocationArgs, true},
		{case1Policy, putObjectActionArgs, true},
		{case1Policy, getObjectActionArgs, true},

		{case2Policy, anonGetBucketLocationArgs, false},
		{case2Policy, anonPutObjectActionArgs, true},
		{case2Policy, anonGetObjectActionArgs, true},
		{case2Policy, getBucketLocationArgs, true},
		{case2Policy, putObjectActionArgs, true},
		{case2Policy, getObjectActionArgs, true},

		{case3Policy, anonGetBucketLocationArgs, false},
		{case3Policy, anonPutObjectActionArgs, true},
		{case3Policy, anonGetObjectActionArgs, false},
		{case3Policy, getBucketLocationArgs, true},
		{case3Policy, putObjectActionArgs, true},
		{case3Policy, getObjectActionArgs, true},

		{case4Policy, anonGetBucketLocationArgs, false},
		{case4Policy, anonPutObjectActionArgs, false},
		{case4Policy, anonGetObjectActionArgs, false},
		{case4Policy, getBucketLocationArgs, true},
		{case4Policy, putObjectActionArgs, false},
		{case4Policy, getObjectActionArgs, true},
	}

	for i, testCase := range testCases {
		result := testCase.policy.IsAllowed(testCase.args)

		if result != testCase.expectedResult {
			t.Fatalf("case %v: expected: %v, got: %v\n", i+1, testCase.expectedResult, result)
		}
	}
}
