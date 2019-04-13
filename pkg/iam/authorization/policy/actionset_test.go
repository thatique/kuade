package iampolicy

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestActionSetMatches(t *testing.T) {
	testCases := []struct {
		actions        ActionSet
		action         Action
		expectedResult bool
	}{
		{NewActionSet(Action("s3:GetObject"), Action("s3:PutObject"), Action("s3:DeleteObject")), Action("s3:GetObject"), true},
		{NewActionSet(Action("s3:*")), Action("s3:PutObject"), true},
	}
	for i, test := range testCases {
		result := test.actions.Match(test.action)

		if result != test.expectedResult {
			t.Fatalf("case %v: expected: %v, got: %v\n", i+1, test.expectedResult, result)
		}
	}
}

func TestActionSetMarshalJSON(t *testing.T) {
	testCases := []struct {
		actionSet      ActionSet
		expectedResult []byte
		expectErr      bool
	}{
		{NewActionSet("s3:PutObject"), []byte(`["s3:PutObject"]`), false},
		{NewActionSet(), nil, true},
	}

	for i, testCase := range testCases {
		result, err := json.Marshal(testCase.actionSet)
		expectErr := (err != nil)

		if expectErr != testCase.expectErr {
			t.Fatalf("case %v: error: expected: %v, got: %v\n", i+1, testCase.expectErr, expectErr)
		}

		if !testCase.expectErr {
			if !reflect.DeepEqual(result, testCase.expectedResult) {
				t.Fatalf("case %v: result: expected: %v, got: %v\n", i+1, string(testCase.expectedResult), string(result))
			}
		}
	}
}
