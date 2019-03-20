package searchie

import (
	"errors"
	"strings"

	"github.com/thatique/kuade/pkg/searchie/tokens"
)

func NewQuery(s string) (terms [][]Term, err error) {
	var t Term
	qs := splitBySpaces(strings.Trim(s, " \t\r\n"))
	var isOr bool
	for _, str := range qs {
		if str == "|" {
			isOr = true
			continue
		} else if str == "&" {
			isOr = false
			continue
		} else {
			isOr = false
		}
		t, err = NewTerm(str)
		if err != nil {
			return
		}
		if isOr {
			terms = conditionOr(terms, free(t))
			continue
		}
		terms = conditionAnd(terms, free(t))
	}
	return
}

func NewTerm(s string) (term Term, err error) {
	toks, err := tokens.Tokenize(s)
	if err != nil {
		return
	}

	p := newSearchParser(toks)
	return p.Parse()
}

type searchParser struct {
	tokens []tokens.Token
	p int
}

func newSearchParser(t []tokens.Token) *searchParser {
	return &searchParser{tokens: t, p: 0}
}

func (p *searchParser) consumeIncluded() bool {
	op := p.p
	_, err := p.when(func(t tokens.Token) bool {
		return t.Kind == tokens.TokenPlus
	})
	if err == nil {
		return true
	}

	p.p = op
	_, err = p.when(func(t tokens.Token) bool {
		return t.Kind == tokens.TokenMinus
	})
	if err == nil {
		return false
	}

	p.p = op
	return true
}

func (p *searchParser) Parse() (term Term, err error) {
	included := p.consumeIncluded()
	var labels []Label
	for {
		op := p.p
		l, err := p.consumeSlabel()
		if err != nil {
			p.p = op
			break
		}
		labels = append(labels, l)
	}

	pred, err := p.consumePredicate()
	if err != nil {
		err = errors.New("error consuming predicate")
		return
	}

	term = Term{Include: included, Labels: labels, Predicate: pred,}
	return
}

func (p *searchParser) consumeText() (str string, err error) {
	tok, err := p.when(tokens.IsTextToken)
	if err != nil {
		return
	}
	str = tok.Text
	return
}

func (p *searchParser) consumeLabel() (label Label, err error) {
	op := p.p
	txt, err := p.consumeText()
	if err != nil {
		p.p = op
		return
	}
	_, err = p.when(func(t tokens.Token) bool {
		return t.Kind == tokens.TokenColon
	})
	if err != nil {
		p.p = op
		return
	}
	label = Label{Common: txt}
	return
}

func (p *searchParser) consumeMeta() (label Label, err error) {
	_, err = p.when(func(t tokens.Token) bool {
		return t.Kind == tokens.TokenAt
	})
	if err != nil {
		return
	}
	lbl, err := p.consumeLabel()
	if err != nil {
		return
	}
	label = Label{Meta: lbl.Common}
	return
}

func (p *searchParser) consumeSlabel() (label Label, err error) {
	op := p.p
	label, err = p.consumeMeta()
	if err == nil {
		return
	}

	p.p = op
	label, err = p.consumeLabel()
	return
}

func (p *searchParser) consumeTag() (val Value, err error) {
	_, err = p.when(func(t tokens.Token) bool {
		return t.Kind == tokens.TokenHash
	})
	if err != nil {
		return
	}
	txt, err := p.consumeText()
	if err != nil {
		return
	}
	val = Value{Tag: txt}
	return
}

func (p *searchParser) consumeValue() (val Value, err error) {
	txt, err := p.consumeText()
	if err != nil {
		return
	}
	val = Value{Text: txt}
	return
}

func (p *searchParser) consumeSValue() (val Value, err error) {
	op := p.p
	val, err = p.consumeTag()
	if err == nil {
		return
	}

	p.p = op
	val, err = p.consumeValue()
	return
}

func (p *searchParser) consumeContains() (pred Predicate, err error) {
	txt, err := p.consumeSValue()
	if err != nil {
		return
	}
	pred = Predicate{Contains: txt}
	return
}

func (p *searchParser) consumeEq() (pred Predicate, err error) {
	_, err = p.when(func(t tokens.Token) bool {
		return t.Kind == tokens.TokenEq
	})
	if err != nil {
		return
	}

	txt, err := p.consumeSValue()
	if err != nil {
		return
	}
	pred = Predicate{Eq: txt}
	return
}

func (p *searchParser) consumeGt() (pred Predicate, err error) {
	_, err = p.when(func(t tokens.Token) bool {
		return t.Kind == tokens.TokenGt
	})
	if err != nil {
		return
	}

	txt, err := p.consumeSValue()
	if err != nil {
		return
	}
	pred = Predicate{Gt: txt}
	return
}

func (p *searchParser) consumeGte() (pred Predicate, err error) {
	_, err = p.when(func(t tokens.Token) bool {
		return t.Kind == tokens.TokenGte
	})
	if err != nil {
		return
	}

	txt, err := p.consumeSValue()
	if err != nil {
		return
	}
	pred = Predicate{Gte: txt}
	return
}

func (p *searchParser) consumeLt() (pred Predicate, err error) {
	_, err = p.when(func(t tokens.Token) bool {
		return t.Kind == tokens.TokenLt
	})
	if err != nil {
		return
	}

	txt, err := p.consumeSValue()
	if err != nil {
		return
	}
	pred = Predicate{Lt: txt}
	return
}

func (p *searchParser) consumeLte() (pred Predicate, err error) {
	_, err = p.when(func(t tokens.Token) bool {
		return t.Kind == tokens.TokenLte
	})
	if err != nil {
		return
	}

	txt, err := p.consumeSValue()
	if err != nil {
		return
	}
	pred = Predicate{Lte: txt}
	return
}

func (p *searchParser) consumeNe() (pred Predicate, err error) {
	_, err = p.when(func(t tokens.Token) bool {
		return t.Kind == tokens.TokenNe
	})
	if err != nil {
		return
	}

	txt, err := p.consumeSValue()
	if err != nil {
		return
	}
	pred = Predicate{Ne: txt}
	return
}

func (p *searchParser) consumeLike() (pred Predicate, err error) {
	_, err = p.when(func(t tokens.Token) bool {
		return t.Kind == tokens.TokenTilde
	})
	if err != nil {
		return
	}

	txt, err := p.consumeText()
	if err != nil {
		return
	}
	pred = Predicate{Like: txt}
	return
}

func (p *searchParser) consumeRange() (pred Predicate, err error) {
	bottom, err := p.consumeSValue()
	if err != nil {
		return
	}
	_, err = p.when(func(t tokens.Token) bool {
		return t.Kind == tokens.TokenRange
	})
	if err != nil {
		return
	}
	top, err := p.consumeSValue()
	if err != nil {
		return
	}
	pred = Predicate{Start: bottom, End: top}
	return
}

func (p *searchParser) consumePredicate() (pred Predicate, err error) {
	op := p.p
	pred, err = p.consumeLike()
	if err == nil {
		return
	}

	p.p = op
	pred, err = p.consumeNe()
	if err == nil {
		return
	}

	p.p = op
	pred, err = p.consumeLte()
	if err == nil {
		return
	}

	p.p = op
	pred, err = p.consumeLt()
	if err == nil {
		return
	}

	p.p = op
	pred, err = p.consumeGt()
	if err == nil {
		return
	}

	p.p = op
	pred, err = p.consumeGte()
	if err == nil {
		return
	}

	p.p = op
	pred, err = p.consumeEq()
	if err == nil {
		return
	}

	p.p = op
	pred, err = p.consumeRange()
	if err == nil {
		return
	}

	p.p = op
	pred, err = p.consumeContains()
	return
}

func (p *searchParser) token() (t tokens.Token, err error) {
	if p.p < len(p.tokens) {
		t = p.tokens[p.p]
		p.p += 1
		return
	}
	err = errors.New("Unexpected EOF")
	return
}

func (p *searchParser) when(pred func(tokens.Token) bool) (t tokens.Token, err error) {
	a, err := p.token()
	if err != nil {
		return
	}
	if pred(a) {
		t = a
		return
	}
	err = errors.New("not satisfy predicate")
	return
}

func splitBySpaces(s string) []string {
	var (
		xs []string
		quoted bool
		buf strings.Builder
	)
	for _, r := range s {
		if r == 34 {
			buf.WriteRune(r)
			quoted = !quoted
			continue
		}

		if r == 32 {
			if !quoted {
				xs = append(xs, buf.String())
				buf.Reset()
				quoted = false
			} else {
				buf.WriteRune(r)
				quoted = true
			}
			continue
		}
		buf.WriteRune(r)
	}

	xs = append(xs, buf.String())

	return xs
}

func free(t Term) [][]Term {
	return [][]Term{[]Term{t}}
}

func conditionAnd(t [][]Term, t2 [][]Term) [][]Term {
	return sliceBind(t, func(xs []Term) [][]Term {
		return sliceBind(t2, func(ys []Term) [][]Term {
			return [][]Term{append(xs, ys...)}
		})
	})
}

func conditionOr(t [][]Term, t2 [][]Term) [][]Term {
	return append(t, t2...)
}

func sliceBind(xs [][]Term, f func([]Term) [][]Term) [][]Term {
	var result [][]Term
	for _, terms := range xs {
		result = append(result, f(terms)...)
	}
	return result
}
