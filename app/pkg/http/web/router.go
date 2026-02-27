package web

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
)

// Request is a wrapper around net/http.Request to introduce additional helper
// method for this package.
type Request struct {
	*http.Request

	log *slog.Logger
}

// Log returns a structured logger that has been contextualized with request
// specific metadata, such as HTTP method, path, request ID etc.
func (r *Request) Log() *slog.Logger {
	return r.log.With(
		slog.String("http_method", r.Method),
		slog.String("http_path", r.URL.Path),
		slog.String("http_host", r.Host),
		slog.String("http_remote_addr", r.RemoteAddr),
		slog.String("http_request_id", GetRequestID(r.Context()).String()),
	)
}

// URLParam retrieves the value of a URL parameter that has been embedded
// within the router using curly brackets (i.e. `/foo/{bar}`). If a param has
// not been set, an empty string will be returned.
func URLParam(ctx context.Context, key string) string {
	return chi.URLParamFromCtx(ctx, key)
}

// Handler is executed in response to a request from a user. It should return
// the template to render, or an error to be handled by ErrorHandler.
type Handler func(context.Context, *Request) (Template, error)

// ErrorHandler is a special Handler which additionally takes the error
// returned by a Handler that could not be handled, it should track this error
// somehow and return a generic error page Template to render.
type ErrorHandler func(context.Context, *Request, error) Template

// Router is an HTTP router that is optimized to handling web applications.
// Internally it contains a Chi router, error handler and logger.
type Router struct {
	r   chi.Router
	err ErrorHandler
	log *slog.Logger
}

// New initializes a new Router, with log being the destination for unhandled
// errors.
func New(log *slog.Logger) *Router {
	return &Router{r: chi.NewRouter(), log: log}
}

func (rt *Router) ErrorHandler(hn ErrorHandler) {
	rt.err = hn
}

// NotFound registers a handler to answer requests for routes that do not
// exist. By default, net/http.NotFound will be used.
func (rt *Router) NotFound(hn Handler) {
	rt.r.NotFound(rt.handle(hn))
}

// MethodNotAllowed registers a handler to answer requests for routes that
// exist but not for the requested method. By default,
// net/http.MethodNotAllowed will be used.
func (rt *Router) MethodNotAllowed(hn Handler) {
	rt.r.MethodNotAllowed(rt.handle(hn))
}

// Use appends one or more HTTP middleware functions to the router, to be
// executed before the request is handled.
func (rt *Router) Use(mw ...func(http.Handler) http.Handler) {
	rt.r.Use(mw...)
}

// Route creates a sub-router of Router, where all requests for a given prefix are
// answered by that Router instance.
func (rt *Router) Route(prefix string, fn func(*Router)) {
	rt.r.Route(prefix, func(r chi.Router) {
		fn(&Router{
			r:   r,
			err: rt.err,
			log: rt.log,
		})
	})
}

// Get registers a new route for the GET method for a Handler under path.
func (rt *Router) Get(path string, hn Handler) {
	rt.r.Get(path, rt.handle(hn))
}

// Head registers a new route for the HEAD method for a Handler under path.
func (rt *Router) Head(path string, hn Handler) {
	rt.r.Head(path, rt.handle(hn))
}

// Put registers a new route for the PUT method for a Handler under the path.
func (rt *Router) Put(path string, hn Handler) {
	rt.r.Put(path, rt.handle(hn))
}

// Post registers a new route for the POST method for a Handler under the path.
func (rt *Router) Post(path string, hn Handler) {
	rt.r.Post(path, rt.handle(hn))
}

// Delete registers a new route for the DELETE method for a Handler under the
// path.
func (rt *Router) Delete(path string, hn Handler) {
	rt.r.Delete(path, rt.handle(hn))
}

// Handle attaches a net/http.Handler to the router.
func (rt *Router) Handle(path string, hn http.Handler) {
	rt.r.Handle(path, hn)
}

// handle builds a generic net/http.HandlerFunc for a Handler, implementing
// error handling, templating, and optionally ContentTyper and StatusCoder. If
// the Template returned by Handler is nil, HTTP 204 No Content will be
// returned.
func (rt *Router) handle(hn Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		req := &Request{Request: r, log: rt.log}
		tpl, err := hn(ctx, req)
		if err != nil {
			tpl = rt.err(ctx, req, err)
		}

		// use `Content-Type` from Template if it implements ContentTyper,
		// otherwise fallback to default.
		if ct, ok := tpl.(ContentTyper); ok {
			value := ct.ContentType()
			if value == "" {
				w.Header().Del("Content-Type")
			} else {
				w.Header().Set("Content-Type", ct.ContentType())
			}
		} else {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		}

		// if Template is nil, inform request there is no content to render.
		if tpl == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// use status code from Template if it implements StatusCoder,
		// otherwise fallback to default.
		status := http.StatusOK
		if sc, ok := tpl.(StatusCoder); ok {
			status = sc.StatusCode()
		}

		// this is special case to set the `Location` header, should consider
		// a more generic interface for Templates to set this, like StatusCoder
		// and ContentTyper do.
		if rdr, ok := tpl.(*redirect); ok {
			w.Header().Set("Location", rdr.location)
		}

		w.WriteHeader(status)

		err = tpl.Render(ctx, w)
		if err != nil {
			rt.log.Error("could not render template", slog.String("error", err.Error()))
		}
	}
}

// ServeHTTP implements the net/http.Handler interface, and hands off a HTTP
// request to Router's internal router.
func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// generate a unique Request ID this this request.
	requestID := uuid.Must(uuid.NewV7())

	// insert the Request ID into the context and the response headers.
	r = r.WithContext(setRequestID(r.Context(), requestID))
	w.Header().Set("Request-Id", requestID.String())

	rt.r.ServeHTTP(w, r)
}
