package spec

import (
	"bytes"
	"github.com/discernhq/devx/pkg/wiretap/internal/internaltools"
	"github.com/tidwall/gjson"
	"html/template"
	"maps"
	"net/http"
	"net/url"
)

func BodyTemplate() *template.Template {
	return template.New("").Delims("${{", "}}")
}

type RequestTemplateVars struct {
	body   *bytes.Buffer
	form   url.Values
	header http.Header
	url    *url.URL
}

func (r *RequestTemplateVars) JSON(path string) any {
	return gjson.Get(r.body.String(), path).Value()
}

func (r *RequestTemplateVars) Header(key string) string {
	return r.header.Get(key)
}

func (r *RequestTemplateVars) Form(key string) string {
	return r.form.Get(key)
}

func (r *RequestTemplateVars) URL() *url.URL {
	return r.url
}

func (r *RequestTemplateVars) Body() string {
	return r.body.String()
}

func (r *RequestTemplateVars) Query(key string) string {
	return r.url.Query().Get(key)
}

func RequestToTemplate(req *http.Request) (*RequestTemplateVars, error) {
	buff, err := internaltools.CloneRequestBodyAsBuffer(req)
	if err != nil {
		return nil, err
	}

	if err = req.ParseForm(); err != nil {
		return nil, err
	}

	return &RequestTemplateVars{
		body:   buff,
		form:   maps.Clone(req.Form),
		url:    req.URL,
		header: req.Header.Clone(),
	}, nil
}
