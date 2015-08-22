package main

import (
	"net/http"
	"strings"
)

func ShowPost(w http.ResponseWriter, p *ProjectedPost) {
	paragraphs := []*HTML{
		H.T("h1", nil, H.Text(p.Title)),
	}
	for _, paragraph := range strings.Split(p.Body, "\n\n") {
		paragraphs = append(paragraphs, H.T("p", nil, H.Text(paragraph)))
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	NewHTMLDocument(p.Title,
		H.T("article",
			H.A("class", "post"),
			paragraphs...,
		),
	).WriteHTML(w, "", "  ")
}
