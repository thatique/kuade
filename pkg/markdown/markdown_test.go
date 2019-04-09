package markdown

import (
	"fmt"
	"html/template"
	"reflect"
	"testing"
)

func print(v interface{}) string {
	value := reflect.ValueOf(v)

	if v == nil {
		return "[nil] nil"
	}

	return fmt.Sprintf("[%s] %v", value.Type(), v)
}

func TestMarkdownSimple(t *testing.T) {
	var inputs = []struct{ input, expected string }{
		{`**Hello World**`, `<p><strong>Hello World</strong></p>`},
		{`[My Link](http://example.com/)`, `<p><a href="http://example.com/">My Link</a></p>`},
		{`![My Image](http://example.com/hello.jpg)`, `<p></p>`},
		{`# Hello World`, `<p>Hello World</p>`},
		{`### Hello World`, `<p>Hello World</p>`},
	}
	for i, input := range inputs {
		output := Simple(input.input)
		expected := template.HTML(input.expected)
		if !reflect.DeepEqual(expected, output) {
			t.Fatalf(`[%d] Equals assertion failed. \n Expected: \n\t\t %s\n Actual: \n\t\t %s`, i, print(expected), print(output))
		}
	}
}
