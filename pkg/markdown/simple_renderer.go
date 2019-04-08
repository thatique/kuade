package markdown

import (
	"bytes"

	"github.com/russross/blackfriday"
)

var _ blackfriday.Renderer = (*simpleRenderer)(nil)

type simpleRenderer struct {
	renderer blackfriday.Renderer
}

func NewSimpleRenderer(flags int) blackfriday.Renderer {
	return simpleRenderer{renderer: blackfriday.HtmlRenderer(flags, "", "")}
}

// Blocklevel callbacks

// BlockCode is the code tag callback.
func (p simpleRenderer) BlockCode(out *bytes.Buffer, text []byte, land string) {
	p.renderer.BlockCode(out, text, land)
}

// BlockQuote is the quote tag callback.
func (p simpleRenderer) BlockQuote(out *bytes.Buffer, text []byte) {
	p.renderer.BlockQuote(out, text)
}

// BlockHtml is the HTML tag callback.
func (p simpleRenderer) BlockHtml(out *bytes.Buffer, text []byte) {
	p.renderer.BlockHtml(out, text)
}

// Header is the header tag callback.
func (p simpleRenderer) Header(out *bytes.Buffer, text func() bool, level int, id string) {
	p.Paragraph(out, text)
}

// HRule is the horizontal rule tag callback.
func (p simpleRenderer) HRule(out *bytes.Buffer) {
	p.renderer.HRule(out)
}

// List is the list tag callback.
func (p simpleRenderer) List(out *bytes.Buffer, text func() bool, flags int) {
	p.renderer.List(out, text, flags)
}

// ListItem is the list item tag callback.
func (p simpleRenderer) ListItem(out *bytes.Buffer, text []byte, flags int) {
	p.renderer.ListItem(out, text, flags)
}

// Paragraph is the paragraph tag callback.  This renders simple paragraph text
// into plain text, such that summaries can be easily generated.
func (p simpleRenderer) Paragraph(out *bytes.Buffer, text func() bool) {
	p.renderer.Paragraph(out, text)
}

// Table is the table tag callback.
func (p simpleRenderer) Table(out *bytes.Buffer, header []byte, body []byte, columnData []int) {
	p.renderer.Table(out, header, body, columnData)
}

// TableRow is the table row tag callback.
func (p simpleRenderer) TableRow(out *bytes.Buffer, text []byte) {
	p.renderer.TableRow(out, text)
}

// TableHeaderCell is the table header cell tag callback.
func (p simpleRenderer) TableHeaderCell(out *bytes.Buffer, text []byte, flags int) {
	p.renderer.TableHeaderCell(out, text, flags)
}

// TableCell is the table cell tag callback.
func (p simpleRenderer) TableCell(out *bytes.Buffer, text []byte, flags int) {
	p.renderer.TableCell(out, text, flags)
}

// Footnotes is the foot notes tag callback.
func (p simpleRenderer) Footnotes(out *bytes.Buffer, text func() bool) {
	p.renderer.Footnotes(out, text)
}

// FootnoteItem is the footnote item tag callback.
func (p simpleRenderer) FootnoteItem(out *bytes.Buffer, name, text []byte, flags int) {
	p.renderer.FootnoteItem(out, name, text, flags)
}

// TitleBlock is the title tag callback.
func (p simpleRenderer) TitleBlock(out *bytes.Buffer, text []byte) {
	p.renderer.TitleBlock(out, text)
}

// Spanlevel callbacks

// AutoLink is the autolink tag callback.
func (p prsimpleRendereroxy) AutoLink(out *bytes.Buffer, link []byte, kind int) {
	p.renderer.AutoLink(out, link, kind)
}

// CodeSpan is the code span tag callback.  Outputs a simple Markdown version
// of the code span.
func (p simpleRenderer) CodeSpan(out *bytes.Buffer, text []byte) {
	p.renderer.CodeSpan(out, text)
}

// DoubleEmphasis is the double emphasis tag callback.  Outputs a simple
// plain-text version of the input.
func (p simpleRenderer) DoubleEmphasis(out *bytes.Buffer, text []byte) {
	p.renderer.DoubleEmphasis(out, text)
}

// Emphasis is the emphasis tag callback.  Outputs a simple plain-text
// version of the input.
func (p simpleRenderer) Emphasis(out *bytes.Buffer, text []byte) {
	p.renderer.Emphasis(out, text)
}

// Image is the image tag callback.
func (p simpleRenderer) Image(out *bytes.Buffer, link []byte, title []byte, alt []byte) {
	p.renderer.Image(out, link, title, alt)
}

// LineBreak is the line break tag callback.
func (p simpleRenderer) LineBreak(out *bytes.Buffer) {
	p.renderer.LineBreak(out)
}

// Link is the link tag callback.  Outputs a simple plain-text version
// of the input.
func (p simpleRenderer) Link(out *bytes.Buffer, link []byte, title []byte, content []byte) {
	p.renderer.Link(out, link, title, content)
}

// RawHtmlTag is the raw HTML tag callback.
func (p simpleRenderer) RawHtmlTag(out *bytes.Buffer, tag []byte) {
	p.renderer.RawHtmlTag(out, tag)
}

// TripleEmphasis is the triple emphasis tag callback.  Outputs a simple plain-text
// version of the input.
func (p simpleRenderer) TripleEmphasis(out *bytes.Buffer, text []byte) {
	p.renderer.TripleEmphasis(out, text)
}

// StrikeThrough is the strikethrough tag callback.
func (p simpleRenderer) StrikeThrough(out *bytes.Buffer, text []byte) {
	p.renderer.StrikeThrough(out, text)
}

// FootnoteRef is the footnote ref tag callback.
func (p simpleRenderer) FootnoteRef(out *bytes.Buffer, ref []byte, id int) {
	p.renderer.FootnoteRef(out, ref, id)
}

// Lowlevel callbacks

// Entity callback.  Outputs a simple plain-text version of the input.
func (p simpleRenderer) Entity(out *bytes.Buffer, entity []byte) {
	p.renderer.Entity(out, entity)
}

// NormalText callback.  Outputs a simple plain-text version of the input.
func (p simpleRenderer) NormalText(out *bytes.Buffer, text []byte) {
	p.renderer.NormalText(out, text)
}

// Header and footer

// DocumentHeader callback.
func (p simpleRenderer) DocumentHeader(out *bytes.Buffer) {
	p.renderer.DocumentHeader(out)
}

// DocumentFooter callback.
func (p simpleRenderer) DocumentFooter(out *bytes.Buffer) {
	p.renderer.DocumentFooter(out)
}

// GetFlags returns zero.
func (p simpleRenderer) GetFlags() int {
	return p.renderer.GetFlags()
}
