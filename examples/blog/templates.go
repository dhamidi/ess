package main

import (
	"net/http"
	"strings"

	"github.com/dhamidi/ess"
)

const stylesheet = `
.menu form {
  display: inline-block;
}

.menu form, .menu form p {
  margin: 0px 5px;
}

.menu a {
  text-decoration: none;
  color: inherit;
}
`

func ShowPost(w http.ResponseWriter, p *ProjectedPost) {
	paragraphs := []*HTML{
		H.T("h1", nil, H.Text(p.Title)),
	}
	for _, paragraph := range p.Paragraphs {
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

func InputField(label, kind, name, value string, errors []string) *HTML {
	wrapper := H.T("label", nil, H.Text(label))
	input := H.T("input",
		H.
			A("type", kind).
			A("value", value).
			A("name", name),
	)

	if kind == "textarea" {
		input = H.T("textarea", H.A("name", name), H.Text(value))
	}

	wrapper.C(input)

	if len(errors) > 0 {
		wrapper.C(H.T("em", H.A("class", "errors"), H.Text(strings.Join(errors, ", "))))
	}

	return wrapper
}

type HTMLFormField struct {
	Name   string
	Kind   string
	Value  string
	Label  string
	Errors []string
}

type HTMLForm struct {
	name   string
	action string

	fields []*HTMLFormField
	params map[string]string

	index map[string]*HTMLFormField
}

func Form(name, action string, fields ...*HTMLFormField) *HTMLForm {
	form := &HTMLForm{
		name:   name,
		action: action,
		fields: fields,
		index:  map[string]*HTMLFormField{},
		params: map[string]string{},
	}

	for _, field := range fields {
		form.index[field.Name] = field
	}

	return form
}

func (self *HTMLForm) Copy() *HTMLForm {
	newFields := make([]*HTMLFormField, len(self.fields))
	for i, field := range self.fields {
		newField := *field
		newFields[i] = &newField
		self.index[field.Name] = &newField
	}
	self.fields = newFields
	return self
}

func (self *HTMLForm) Action(action string) *HTMLForm {
	self.action = action
	return self
}

func (self *HTMLForm) Param(name, value string) *HTMLForm {
	self.params[name] = value
	return self
}

func (self *HTMLForm) Fill(params *ess.Command, err error) *HTMLForm {
	self.Copy()
	verr, hasErrors := err.(*ess.ValidationError)

	for name, value := range params.Fields {
		field, found := self.index[name]
		if !found {
			continue
		}

		if hasErrors {
			field.Errors = verr.Errors[name]
		}

		if field.Kind != "password" {
			field.Value = value.String()
		}
	}

	return self
}

func (self *HTMLForm) ToHTML(submit string) *HTML {
	rows := []*HTML{}

	self.addParams(&rows)

	for _, field := range self.fields {
		row := H.T("p", nil, InputField(field.Label, field.Kind, field.Name, field.Value, field.Errors))
		rows = append(rows, row)
	}

	rows = append(rows,
		H.T("p", nil,
			H.T("button", H.A("type", "submit"),
				H.Text(submit),
			),
		),
	)

	return H.T("form",
		H.
			A("id", self.name).
			A("action", self.action).
			A("method", "POST"),
		rows...,
	)
}

func (self *HTMLForm) addParams(rows *[]*HTML) {
	if len(self.params) == 0 {
		return
	}

	row := H.T("p", nil)
	for param, value := range self.params {
		row.C(H.T("input",
			H.
				A("type", "hidden").
				A("name", param).
				A("value", value),
		))
	}

	*rows = append(*rows, row)
}

var (
	SignUpForm = Form("signup", "/signups",
		&HTMLFormField{Label: "Username", Name: "username", Kind: "text"},
		&HTMLFormField{Label: "Email", Name: "email", Kind: "email"},
		&HTMLFormField{Label: "Password", Name: "password", Kind: "password"},
	)

	LoginForm = Form("login", "/sessions",
		&HTMLFormField{Label: "Username", Name: "username", Kind: "text"},
		&HTMLFormField{Label: "Password", Name: "password", Kind: "password"},
	)

	LogoutForm = Form("logout", "/sessions/logout")

	PostForm = Form("write-post", "/posts",
		&HTMLFormField{Label: "Path", Name: "id", Kind: "text"},
		&HTMLFormField{Label: "Title", Name: "title", Kind: "text"},
		&HTMLFormField{Label: "Body", Name: "body", Kind: "textarea"},
	)

	EditPostForm = Form("edit-post", "/posts/edit",
		&HTMLFormField{Label: "Reason", Name: "reason", Kind: "text"},
		&HTMLFormField{Label: "Title", Name: "title", Kind: "text"},
		&HTMLFormField{Label: "Body", Name: "body", Kind: "textarea"},
	)
)

func ShowSignupForm(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	NewHTMLDocument("Sign up",
		SignUpForm.ToHTML("Sign up"),
	).WriteHTML(w, "", "  ")
}

func ShowSignupFormErrors(w http.ResponseWriter, params *ess.Command, err error) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	NewHTMLDocument("Sign up",
		SignUpForm.Fill(params, err).ToHTML("Sign up"),
	).WriteHTML(w, "", "  ")
}

func ShowLoginForm(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	NewHTMLDocument("Log In",
		LoginForm.Copy().Param("return", req.FormValue("return")).ToHTML("Log in"),
	).WriteHTML(w, "", "  ")
}

func ShowLoginFormError(w http.ResponseWriter, params *ess.Command, err error) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	NewHTMLDocument("Log In",
		LoginForm.Fill(params, err).ToHTML("Log in"),
	).WriteHTML(w, "", "  ")

}

func PostOnIndex(post *ProjectedPost, currentUser *ProjectedUser) *HTML {
	result := H.T("article", nil,
		H.T("em", nil,
			H.Text(post.WrittenAt.Format("_2 Jan 2006 "))),
		H.Text("by "),
		H.T("em", nil,
			H.Text(post.Author)),
		H.T("a",
			H.A("href", post.Path),
			H.Text(post.Title),
		),
	)

	if currentUser != nil && post.Author == currentUser.Username {
		result.C(
			H.T("a", H.A("href", post.Path+"/edit"),
				H.T("button", nil, H.Text("Edit"))),
		)
	}

	result.C(
		H.T("blockquote", nil,
			H.T("p", nil, H.Text(post.Paragraphs[0]))),
	)

	return result
}

func ShowAllPostsIndex(w http.ResponseWriter, currentUser *ProjectedUser, posts []*ProjectedPost) {
	menu := H.T("div", H.A("class", "menu"))
	if currentUser == nil {
		menu.C(
			H.T("a", H.A("href", "/sessions"),
				H.T("button", nil, H.Text("Log in"))),
			H.T("a", H.A("href", "/signups"),
				H.T("button", nil, H.Text("Sign up"))),
		)
	} else {
		menu.C(
			H.T("a", H.A("href", "/posts"),
				H.T("button", nil, H.Text("Write post"))),
			LogoutForm.ToHTML("Log out"),
		)
	}

	body := H.T("ul", nil)
	for _, post := range posts {
		body.C(H.T("li", nil, PostOnIndex(post, currentUser)))
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	NewHTMLDocument("Recent posts",
		menu,
		H.T("h1", nil, H.Text("Recent posts")),
		body,
	).WriteHTML(w, "", "  ")
}

func ShowPostForm(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	NewHTMLDocument("Write Post",
		PostForm.ToHTML("Write post"),
	).WriteHTML(w, "", "  ")

}

func ShowPostFormError(w http.ResponseWriter, params *ess.Command, err error) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	NewHTMLDocument("Write Post",
		PostForm.Fill(params, err).ToHTML("Write post"),
	).WriteHTML(w, "", "  ")

}

func ShowEditPostForm(w http.ResponseWriter, params *ess.Command) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	NewHTMLDocument("Edit Post",
		EditPostForm.
			Fill(params, nil).
			Action("/posts/"+params.AggregateId()+"/edit").
			Param("id", params.AggregateId()).
			Param("username", params.Get("username").String()).
			ToHTML("Edit post"),
	).WriteHTML(w, "", "  ")

}

func ShowEditPostFormError(w http.ResponseWriter, params *ess.Command, err error) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	NewHTMLDocument("Edit Post",
		EditPostForm.
			Fill(params, err).
			Action("/posts/"+params.AggregateId()+"/edit").
			Param("id", params.AggregateId()).
			Param("username", params.Get("username").String()).
			ToHTML("Edit post"),
	).WriteHTML(w, "", "  ")

}
