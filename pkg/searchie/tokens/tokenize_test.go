package tokens

import (
	"testing"
)

var inputs = []string{
	">2", "foo", "+foo", "-foo", "#foo", "*", "uni*", "foo:>2", "foo:0..2",
	"-foo:0..2", "foo:bar:baz", "baz:~\"_foo%bar\"",
	"~?foo*bar", "foo:uni*", "@path:/foo/bar", "path:\"foo bar\"",
}

func TestTokenize(t *testing.T) {
	for i, inpt := range inputs {
		tokens, err := Tokenize(inpt)
		if err != nil {
			t.Fatalf("%d: %v", i, err)
		}

		if len(tokens) == 0 {
			t.Fatal("expected non empty slice")
		}
	}
}
