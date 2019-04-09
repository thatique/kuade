package markdown

import (
	"io"

	"github.com/russross/blackfriday"
)

var _ blackfriday.Renderer = SimpleRenderer{}

type SimpleRenderer struct {
	blackfriday.Renderer
}

func NewSimpleRenderer(flags blackfriday.HTMLFlags) blackfriday.Renderer {
	return SimpleRenderer{Renderer: blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		Flags: flags,
	})}
}

func (p SimpleRenderer) RenderNode(w io.Writer, node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
	if node.Type == blackfriday.Heading {
		node.Type = blackfriday.Paragraph
		return p.Renderer.RenderNode(w, node, entering)
	}
	// otherwise call original one
	return p.Renderer.RenderNode(w, node, entering)
}
