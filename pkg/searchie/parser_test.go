package searchie

import (
	"testing"
)

func TestNewTerm(t *testing.T) {
	var inputs = []string{
		">2", "foo", "+foo", "-foo", "#foo", "*", "uni*", "foo:>2", "foo:0..2",
		"-foo:0..2", "foo:bar:baz", "baz:~\"_foo%bar\"",
		"~?foo*bar", "foo:uni*", "@path:/foo/bar", "path:\"foo bar\"",
	}
	for i, inpt := range inputs {
		_, err := NewTerm(inpt)
		if err != nil {
			t.Fatalf("%d: %v", i, err)
		}
	}
}
