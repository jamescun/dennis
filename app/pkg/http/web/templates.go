package web

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

// Template is implemented by types that can be marshaled and written in
// response to a user request.
//
// This is compatible with templ.Component.
type Template interface {
	Render(context.Context, io.Writer) error
}

// StatusCoder is optionally implemented by Templates to control the HTTP
// status code returned. If not implemented, HTTP 200 OK will be used.
type StatusCoder interface {
	StatusCode() int
}

// ContentTyper is optionally implemented by Templates to control the HTTP
// `Content-Type` header. If not implemented, `text/html; charset=utf-8` will
// be used. If the value returned by ContentType() is an empty string, the
// `Content-Type` header will be removed.
type ContentTyper interface {
	ContentType() string
}

// jsonTemplate is a Template wrapper that renders an object as JSON as if it
// were a Template.
type jsonTemplate struct {
	src any
}

func (j *jsonTemplate) ContentType() string {
	return "application/json; charset=utf-8"
}

func (j *jsonTemplate) StatusCode() int {
	if sc, ok := j.src.(StatusCoder); ok {
		return sc.StatusCode()
	}

	return http.StatusOK
}

func (j *jsonTemplate) Render(_ context.Context, w io.Writer) error {
	return json.NewEncoder(w).Encode(j.src)
}

// JSON is a Template wrapper that will render src as a JSON object, setting
// the relevant HTTP `Content-Type` header, and optionally supporting the
// StatusCoder interface if on src.
func JSON(src any) Template {
	return &jsonTemplate{src: src}
}

// redirect is a special Template implementation that redirects the request
// somewhere else, à la net/http.Redirect().
type redirect struct {
	location string
	status   int
}

func (r *redirect) StatusCode() int {
	// if status is not configured for some reason, default to HTTP See Other.
	if r.status < 0 {
		return http.StatusSeeOther
	}

	return r.status
}

func (r *redirect) Render(_ context.Context, w io.Writer) error {
	// write dumb page with click to redirect for compatibility with whatever
	// doesn't respect the `Location` header.
	_, err := w.Write([]byte(`You are being redirected, <a href="` + r.location + `">Click Here</a> if you are not.`))
	if err != nil {
		return err
	}

	return nil
}

// Redirect is a special component that signals to Web to redirect the request
// rather than render a page. A basic page will be sent to the user informing
// them they are being redirected if for whatever reason the `Location` header
// does not get respected.
func Redirect(location string, status int) Template {
	return &redirect{location: location, status: status}
}
