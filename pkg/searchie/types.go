package searchie

import (
	"fmt"
	"strings"
)

type Term struct {
	Include   bool
	Labels    []Label
	Predicate Predicate
}

func (t Term) Show() string {
	var labels []string
	for _, l := range t.Labels {
		labels = append(labels, l.Show())
	}

	return fmt.Sprintf(
		"Term{Labels: [%s], Include: %v, Predicate: %v",
		strings.Join(labels, ", "),
		t.Include,
		t.Predicate.Show())
}

type Label struct {
	Common, Meta string
}

func (label Label) Show() string {
	if label.Common != "" {
		return fmt.Sprintf("Common{%s}", label.Common)
	}

	return fmt.Sprintf("Meta{%s}", label.Meta)
}

type Value struct {
	Text string
	Tag  string
}

func (val Value) IsText() bool {
	return val.Text != ""
}

func (val Value) IsTag() bool {
	return val.Tag != ""
}

func (val Value) String() string {
	if val.Tag != "" {
		return val.Tag
	}

	return val.Text
}

func (val Value) Show() string {
	if val.Tag != "" {
		return fmt.Sprintf("Tag{#%s}", val.Tag)
	}

	return fmt.Sprintf("Text{%s}", val.Text)
}

type Predicate struct {
	Contains, Eq, Gt, Gte, Lt, Lte, Ne Value

	Like string

	// Range
	Start, End Value
}

func (p Predicate) Show() string {
	if p.Contains.String() != "" {
		return fmt.Sprintf("Contains{%s}", p.Contains.Show())
	}

	if p.Eq.String() != "" {
		return fmt.Sprintf("Eq{%s}", p.Eq.Show())
	}

	if p.Gt.String() != "" {
		return fmt.Sprintf("Gt{%s}", p.Gt.Show())
	}

	if p.Gte.String() != "" {
		return fmt.Sprintf("Gte{%s}", p.Gte.Show())
	}

	if p.Lt.String() != "" {
		return fmt.Sprintf("Lt{%s}", p.Lt.Show())
	}

	if p.Lte.String() != "" {
		return fmt.Sprintf("Lte{%s}", p.Lte.Show())
	}

	if p.Ne.String() != "" {
		return fmt.Sprintf("Ne{%s}", p.Ne.Show())
	}

	if p.Like != "" {
		return fmt.Sprintf("Like{%s}", p.Like)
	}

	return fmt.Sprintf("Range{%s, %s}", p.Start.Show(), p.End.Show())
}
