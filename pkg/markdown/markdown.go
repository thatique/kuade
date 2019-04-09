package markdown

import (
	"html"
	"html/template"
	"strings"

	"github.com/russross/blackfriday"
)

var mdExtns = 0 |
	blackfriday.Tables |
	blackfriday.Autolink |
	blackfriday.FencedCode |
	blackfriday.Titleblock |
	blackfriday.Strikethrough |
	blackfriday.DefinitionLists |
	blackfriday.NoIntraEmphasis |
	blackfriday.HardLineBreak

var simpleHTMLExtensions = 0 |
	blackfriday.UseXHTML |
	blackfriday.Smartypants |
	blackfriday.SmartypantsFractions |
	blackfriday.SmartypantsDashes |
	blackfriday.SkipImages |
	blackfriday.SmartypantsLatexDashes

var simpleRenderer = NewSimpleRenderer(simpleHTMLExtensions)

// Simple turns a markdown into HTML using few rules
func Simple(input string) template.HTML {
	sanitizedInput := html.EscapeString(input)
	output := blackfriday.Run([]byte(sanitizedInput),
		blackfriday.WithRenderer(simpleRenderer),
		blackfriday.WithExtensions(mdExtns),
	)

	return template.HTML(strings.TrimSpace(string(output)))
}

var fullHTMLExtensions = 0 |
	blackfriday.UseXHTML |
	blackfriday.Smartypants |
	blackfriday.SmartypantsFractions |
	blackfriday.SmartypantsDashes |
	blackfriday.SmartypantsLatexDashes

var fullRenderer = blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
	Flags: fullHTMLExtensions,
})

// Full turns a markdown into HTML using all rules
func Full(input string) template.HTML {
	sanitizedInput := html.EscapeString(input)
	output := blackfriday.Run([]byte(sanitizedInput),
		blackfriday.WithRenderer(fullRenderer),
		blackfriday.WithExtensions(mdExtns),
	)

	return template.HTML(strings.TrimSpace(string(output)))
}
