package iam

import "testing"

func TestIsAccessKeyValid(t *testing.T) {
	testCases := []struct {
		accessKey      string
		expectedResult bool
	}{
		{AccessKeyChars[:accessKeyMinLen], true},
		{AccessKeyChars[:accessKeyMinLen+1], true},
		{AccessKeyChars[:accessKeyMinLen-1], false},
	}

	for i, testCase := range testCases {
		result := IsAccessKeyValid(testCase.accessKey)
		if result != testCase.expectedResult {
			t.Fatalf("test %v: expected: %v, got: %v", i+1, testCase.expectedResult, result)
		}
	}
}

func TestIsSecretKeyValid(t *testing.T) {
	testCases := []struct {
		secretKey      string
		expectedResult bool
	}{
		{AccessKeyChars[:secretKeyMinLen], true},
		{AccessKeyChars[:secretKeyMinLen+1], true},
		{AccessKeyChars[:secretKeyMinLen-1], false},
	}

	for i, testCase := range testCases {
		result := IsSecretKeyValid(testCase.secretKey)
		if result != testCase.expectedResult {
			t.Fatalf("test %v: expected: %v, got: %v", i+1, testCase.expectedResult, result)
		}
	}
}

func TestGetNewAccessKey(t *testing.T) {
	acc, err := GetNewAccessKey()
	if err != nil {
		t.Fatalf("Failed to get a new credential")
	}
	if !acc.IsValid() {
		t.Fatalf("Failed to get new valid credential")
	}
	if len(acc.AccessKey) != accessKeyMaxLen {
		t.Fatalf("access key length: expected: %v, got: %v", accessKeyMaxLen, len(acc.AccessKey))
	}
	if len(acc.SecretKey) != secretKeyMaxLen {
		t.Fatalf("secret key length: expected: %v, got: %v", secretKeyMaxLen, len(acc.SecretKey))
	}
}
