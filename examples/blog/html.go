package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
)

type HTML struct {
	Name       string
	Attributes HTMLAttributes
	Children   []*HTML
	Content    string
}

var (
	escape = template.HTMLEscapeString
	H      = &HTML{}
)

func (self *HTML) String() string {
	buf := new(bytes.Buffer)
	self.WriteHTML(buf, "", "  ")
	return buf.String()
}

func (self *HTML) WriteHTML(w io.Writer, prefix, indent string) {
	if self.Name == "TEXT" {
		fmt.Fprintf(w, "%s%s", prefix, escape(self.Content))
		return
	}

	out := w
	if out == nil {
		out = new(bytes.Buffer)
	}

	fmt.Fprintf(out, "%s<%s", prefix, self.Name)
	if len(self.Attributes) > 0 {
		fmt.Fprintf(out, " ")
		count := 0
		for attr, value := range self.Attributes {
			fmt.Fprintf(out, `%s="%s"`, escape(attr), escape(value))
			count++
			if count < len(self.Attributes) {
				fmt.Fprintf(out, " ")
			}
		}
	}
	fmt.Fprintf(out, ">\n")
	for _, child := range self.Children {
		child.WriteHTML(out, prefix+indent, indent)
		fmt.Fprintf(out, "\n")
	}
	fmt.Fprintf(out, "%s</%s>", prefix, self.Name)
}

func (self *HTML) T(name string, attributes HTMLAttributes, children ...*HTML) *HTML {
	return &HTML{
		Name:       name,
		Attributes: attributes,
		Children:   children,
	}
}

func (self *HTML) Text(text string) *HTML {
	return &HTML{Content: text, Name: "TEXT"}
}

func (self *HTML) A(name, value string) HTMLAttributes {
	return HTMLAttributes{name: value}
}

type HTMLAttributes map[string]string

func (self HTMLAttributes) A(name, value string) HTMLAttributes {
	self[name] = value
	return self
}

func NewHTMLDocument(title string, body ...*HTML) *HTML {
	h := &HTML{}
	return h.T("html", nil,
		h.T("head", nil,
			h.T("meta", h.A("charset", "utf")),
			h.T("title", nil, h.Text(title)),
		),
		h.T("body", nil, body...),
	)
}
