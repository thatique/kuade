package template

import (
	"crypto/md5"
	"html/template"
	"io"
	"path"
	"path/filepath"

	"github.com/thatique/kuade/pkg/markdown"
)

var templateFunctions = template.FuncMap{
	"md5": func(input string) string {
		hash := md5.New()
		hash.Write([]byte(data))
		return hash.Sum(nil)
	},
	"markdown": func(input string) template.HTML {
		return markdown.Full(input)
	},
}

type M map[string]interface{}

type Renderer struct {
	// we will load the template from this function
	assets   func(string) ([]byte, error)
	// the cached template
	templates map[string]*template.Template
}

func (r *Renderer) Render(w io.Writer, props interface{}, name, tpls ...string) {
	tpl, err := r.Template(name, tpls...)
	if err != nil {
		panic(err)
	}
	if err = tpl.Execute(w, props); err != nil {
		panic(err)
	}
}

func (r *Renderer) Template(name string, tpls ...string) (tpl *template.Template, err error) {
	if t, ok := r.templates[name]; ok {
		return t, nil
	}

	tpl = template.New(name).Func(templateFunctions)
	for _, tn := range tpls {
		if tpl, err = r.parseTemplate(tpl, tn); err != nil {
			return nil, err
		}
	}

	r.templates[name] = tpl

	return
}

func (r *Renderer) parse(tpl *template.Template, name string) (*template.Template, error) {
	assetPath := getTemplatePath(name)
	var (
		b []byte
		err error
	)
	if b, err := r.assets(assetPath); err != nil {
		return nil, err
	}
	return tpl.Parse(string(b))
}

func getTemplatePath(name string) string {
	assetPath := path.Join("assets/templates", filepath.FromSlash(path.Clean("/"+name)))
	if len(assetPath) > 0 && assetPath[0] == '/' {
		assetPath = assetPath[1:]
	}
	return assetPath
}
