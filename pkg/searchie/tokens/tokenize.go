package tokens

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

type TokenType int

const (
	TokenText TokenType = iota
	TokenRange
	TokenHash
	TokenPlus
	TokenMinus
	TokenAt
	TokenEq
	TokenLt
	TokenGt
	TokenLte
	TokenGte
	TokenNe
	TokenTilde
	TokenColon
)

func Tokenize(s string) ([]Token, error) {
	p := newTokenParser(s)
	return p.parse()
}

type Token struct {
	Kind TokenType
	Text string
}

func IsTextToken(t Token) bool {
	return t.Kind == TokenText
}

func (t Token) Show() string {
	switch t.Kind {
	case TokenText:
		return fmt.Sprintf("TokenText{%s}", t.Text)
	case TokenRange:
		return "TokenRange"
	case TokenHash:
		return "TokenHash"
	case TokenPlus:
		return "TokenPlus"
	case TokenMinus:
		return "TokenMinus"
	case TokenAt:
		return "TokenAt"
	case TokenEq:
		return "TokenEq"
	case TokenLt:
		return "TokenLt"
	case TokenGt:
		return "TokenGt"
	case TokenLte:
		return "TokenLte"
	case TokenGte:
		return "TokenGte"
	case TokenNe:
		return "TokenNe"
	case TokenTilde:
		return "TokenTilde"
	case TokenColon:
		return "TokenColon"
	default:
		return "invalid"
	}
}

type tokenParser struct {
	s string
}

func newTokenParser(s string) *tokenParser {
	return &tokenParser{
		s: s,
	}
}

func (p *tokenParser) parse() (tokens []Token, err error) {
	var tok Token
	for p.len() > 0 {
		tok, err = p.tokenize()
		if err != nil {
			break
		}
		tokens = append(tokens, tok)
	}
	return
}

func (p *tokenParser) tokenize() (t Token, err error) {
	p2 := *p
	t, err = p.colon()
	if err == nil {
		return t, err
	}

	*p = p2
	p2 = *p
	t, err = p.tilde()
	if err == nil {
		return t, err
	}

	*p = p2
	p2 = *p
	t, err = p.ne()
	if err == nil {
		return t, err
	}

	*p = p2
	p2 = *p
	t, err = p.gte()
	if err == nil {
		return t, err
	}

	*p = p2
	p2 = *p
	t, err = p.lte()
	if err == nil {
		return t, err
	}

	*p = p2
	p2 = *p
	t, err = p.gt()
	if err == nil {
		return t, err
	}

	*p = p2
	p2 = *p
	t, err = p.lt()
	if err == nil {
		return t, err
	}

	*p = p2
	p2 = *p
	t, err = p.eq()
	if err == nil {
		return t, err
	}

	*p = p2
	p2 = *p
	t, err = p.at()
	if err == nil {
		return t, err
	}

	*p = p2
	p2 = *p
	t, err = p.minus()
	if err == nil {
		return t, err
	}

	*p = p2
	p2 = *p
	t, err = p.plus()
	if err == nil {
		return t, err
	}

	*p = p2
	p2 = *p
	t, err = p.hash()
	if err == nil {
		return t, err
	}

	*p = p2
	p2 = *p
	t, err = p.ranged()
	if err == nil {
		return t, err
	}

	*p = p2
	p2 = *p
	t, err = p.quoted()
	if err == nil {
		return t, err
	}

	*p = p2
	t, err = p.raw()
	if err == nil {
		return t, err
	}

	return Token{}, err
}

func (p *tokenParser) raw() (tok Token, err error) {
	raw, err := p.rawString()
	if err != nil {
		return Token{}, err
	}

	return Token{Kind: TokenText, Text: raw,}, nil
}

func (p *tokenParser) quoted() (tok Token, err error) {
	raw, err := p.quotedString()
	if err != nil {
		return Token{}, err
	}

	return Token{Kind: TokenText, Text: raw,}, nil
}

func (p *tokenParser) ranged() (tok Token, err error) {
	if p.strings("..") {
		return Token{Kind: TokenRange}, nil
	}
	return Token{}, errors.New("not range token")
}

func (p *tokenParser) hash() (tok Token, err error) {
	if p.consume('#') {
		return Token{Kind: TokenHash}, nil
	}
	return Token{}, errors.New("not hash token")
}

func (p *tokenParser) plus() (tok Token, err error) {
	if p.consume('+') {
		return Token{Kind: TokenPlus}, nil
	}
	return Token{}, errors.New("not plus token")
}

func (p *tokenParser) minus() (tok Token, err error) {
	if p.consume('-') {
		return Token{Kind: TokenMinus}, nil
	}
	return Token{}, errors.New("not range token")
}

func (p *tokenParser) at() (tok Token, err error) {
	if p.consume('@') {
		return Token{Kind: TokenAt}, nil
	}
	return Token{}, errors.New("not range token")
}

func (p *tokenParser) eq() (tok Token, err error) {
	if p.consume('=') {
		return Token{Kind: TokenEq}, nil
	}
	return Token{}, errors.New("not range token")
}

func (p *tokenParser) lt() (tok Token, err error) {
	if p.consume('<') {
		return Token{Kind: TokenLt}, nil
	}
	return Token{}, errors.New("not range token")
}

func (p *tokenParser) gt() (tok Token, err error) {
	if p.consume('>') {
		return Token{Kind: TokenGt}, nil
	}
	return Token{}, errors.New("not range token")
}

func (p *tokenParser) lte() (tok Token, err error) {
	if p.strings("<=") {
		return Token{Kind: TokenLte}, nil
	}
	return Token{}, errors.New("not range token")
}

func (p *tokenParser) gte() (tok Token, err error) {
	if p.strings(">=") {
		return Token{Kind: TokenGte}, nil
	}
	return Token{}, errors.New("not range token")
}

func (p *tokenParser) ne() (tok Token, err error) {
	p2 := *p
	if p.strings("!=") {
		return Token{Kind: TokenNe}, nil
	}
	if p2.strings("<>") {
		return Token{Kind: TokenNe}, nil
	}
	return Token{}, errors.New("not ne token")
}

func (p *tokenParser) tilde() (tok Token, err error) {
	if p.consume('~') {
		return Token{Kind: TokenTilde}, nil
	}
	return Token{}, errors.New("not range token")
}

func (p *tokenParser) colon() (tok Token, err error) {
	if p.consume(':') {
		return Token{Kind: TokenColon}, nil
	}
	return Token{}, errors.New("not colon token")
}

func (p *tokenParser) rawString() (raw string, err error) {
	i := 0
Loop:
	for {
		r, size := utf8.DecodeRuneInString(p.s[i:])
		switch {
		case size == 1 && r == utf8.RuneError:
			return "", fmt.Errorf("incorrect raw string: %q", p.s)

		case size == 0:
			break Loop

		case size == 1 && !IsKeywords(r):
			i += size

		default:
			break Loop
		}
	}

	if i == 0 {
		return "",  fmt.Errorf("empty raw string: %q", p.s)
	}

	raw, p.s = p.s[:i], p.s[i:]

	return raw, nil
}

func (p *tokenParser) quotedSymbol() (b byte, err error) {
	p2 := *p

	b, err = p.slashed()
	if err != nil {
		b = p2.peek()
		*p = p2
		if b != 34 {
			p.s = p.s[1:]
			err = nil
			return
		}

		err = errors.New("end of quoted symbol")
	}

	return
}

func (p *tokenParser) quotedString() (str string, err error) {
	if !p.consume(34) {
		return "", errors.New("not a quoted string")
	}
	var buf strings.Builder
	buf.WriteByte('"')
	for p.len() > 0 {
		b, err := p.quotedSymbol()
		if err != nil {
			break
		}
		buf.WriteByte(b)
	}
	if !p.consume(34) {
		return "", errors.New("invalid quoted string")
	}

	buf.WriteByte('"')

	return buf.String(), nil
}

func (p *tokenParser) slashed() (r byte, err error) {
	if p.consume('\\') {
		r := p.peek()
		p.s = p.s[1:]
		return r, nil
	}

	return r, errors.New("expected slashed string")
}

func (p *tokenParser) consume(c byte) bool {
	if p.empty() || p.peek() != c {
		return false
	}
	p.s = p.s[1:]
	return true
}

func (p *tokenParser) strings(s string) bool {
	if strings.Index(p.s, s) == 0 {
		p.s = p.s[len(s):]
		return true
	}
	return false
}

func (p *tokenParser) peek() byte {
	return p.s[0]
}

func (p *tokenParser) empty() bool {
	return p.len() == 0
}

func (p *tokenParser) len() int {
	return len(p.s)
}

func IsKeywords(r rune) bool {
	switch r {
	case '.', '~', '!', '@', '#', '(', ')', '-', '+', '=', '<', '>', ' ', '"', ':', '|', '&':
		return true
	default:
		return isWhitespace(r)
	}
}

func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n'
}
