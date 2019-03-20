package tokens

import (
	"reflect"
	"strings"
	"testing"
)

var inputs = []string{
	">2", "foo", "+foo", "-foo", "#foo", "*", "uni*", "foo:>2", "foo:0..2",
	"-foo:0..2", "foo:bar:baz", "baz:~\"_foo%bar\"",
	"~?foo*bar", "foo:uni*", "@path:/foo/bar", "path:\"foo bar\"",
}

func showTokens(tokens []Token) string {
	var xs []string
	for _, tok := range tokens {
		xs = append(xs, tok.Show())
	}
	return strings.Join(xs, ", ")
}

func TestTokenize(t *testing.T) {
	var cases = []struct{
		input  string
		tokens []Token
		valid  bool
	}{
		{
			input: ">foo",
			tokens: []Token{
				Token{
					Kind: TokenGt,
				},
				Token{
					Kind: TokenText,
					Text: "foo",
				},
			},
			valid: true,
		},
		{
			input: "foo",
			tokens: []Token{
				Token{
					Kind: TokenText,
					Text: "foo",
				},
			},
			valid: true,
		},
		{
			input: "+foo",
			tokens: []Token{
				Token{
					Kind: TokenPlus,
				},
				Token{
					Kind: TokenText,
					Text: "foo",
				},
			},
			valid: true,
		},
		{
			input: "-foo",
			tokens: []Token{
				Token{
					Kind: TokenMinus,
				},
				Token{
					Kind: TokenText,
					Text: "foo",
				},
			},
			valid: true,
		},
		{
			input: "#foo",
			tokens: []Token{
				Token{
					Kind: TokenHash,
				},
				Token{
					Kind: TokenText,
					Text: "foo",
				},
			},
			valid: true,
		},
		{
			input: "*",
			tokens: []Token{
				Token{
					Kind: TokenText,
					Text: "*",
				},
			},
			valid: true,
		},
		{
			input: "uni*",
			tokens: []Token{
				Token{
					Kind: TokenText,
					Text: "uni*",
				},
			},
			valid: true,
		},
		{
			input: "foo:>2",
			tokens: []Token{
				Token{
					Kind: TokenText,
					Text: "foo",
				},
				Token{
					Kind: TokenColon,
				},
				Token{
					Kind: TokenGt,
				},
				Token{
					Kind: TokenText,
					Text: "2",
				},
			},
			valid: true,
		},
		{
			input: "foo:1..10",
			tokens: []Token{
				Token{
					Kind: TokenText,
					Text: "foo",
				},
				Token{
					Kind: TokenColon,
				},
				Token{
					Kind: TokenText,
					Text: "1",
				},
				Token{
					Kind: TokenRange,
				},
				Token{
					Kind: TokenText,
					Text: "10",
				},
			},
			valid: true,
		},
	}
	for i, cs := range cases {
		tokens, err := Tokenize(cs.input)
		if err != nil && cs.valid {
			t.Fatalf("expected valid token at %d: %v", i, err)
		}
		if !reflect.DeepEqual(cs.tokens, tokens) {
			t.Fatalf("expected tokens: %v is not equal with actual: %v", showTokens(cs.tokens), showTokens(tokens))
		}
	}
}
