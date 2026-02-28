package app

import (
	"context"
	"log/slog"
	"net/http"

	apiv1 "github.com/jamescun/dennis/api/v1"
	"github.com/jamescun/dennis/app/pkg/http/web"
	"github.com/jamescun/dennis/app/views/templates"
)

// UI implements the HTML-based graphical user interface of DENNIS for users to
// interact with.
type UI struct {
	api apiv1.API
	log *slog.Logger
}

// NewUI initializes a new user interface for a given logic backup implementing
// API, and a logger for error messages.
func NewUI(backend apiv1.API, log *slog.Logger) *UI {
	return &UI{
		api: backend,
		log: log,
	}
}

// Routes applies the path-based routes of UI to an HTTP router.
func (ui *UI) Routes(r *web.Router) {
	r.NotFound(ui.NotFound)
	r.ErrorHandler(ui.ErrorHandler)

	r.Get("/", ui.Index)
	r.Post("/query", ui.Query)
	r.Get("/query/{id}", ui.GetQuery)

	// mount the embedded assets for templates.
	r.Handle("/assets/*", templates.Assets("/assets"))
}

func (ui *UI) Index(ctx context.Context, r *web.Request) (web.Template, error) {
	return templates.Index(nil), nil
}

func (ui *UI) Query(ctx context.Context, r *web.Request) (web.Template, error) {
	res, err := ui.api.CreateQuery(ctx, &apiv1.CreateQueryRequest{
		Type: r.FormValue("type"),
		Name: r.FormValue("name"),
	})
	if err != nil {
		if err, ok := err.(*apiv1.Error); ok {
			// a validation error was discovered at the logic layer, display it to
			// the user to try again.
			return templates.Index(err), nil
		}
		return nil, err
	}

	return web.Redirect("/query/"+res.Query.ID.String(), http.StatusSeeOther), nil
}

func (ui *UI) GetQuery(ctx context.Context, r *web.Request) (web.Template, error) {
	res, err := ui.api.GetQuery(ctx, &apiv1.GetQueryRequest{
		ID: web.URLParam(ctx, "id"),
	})
	if err != nil {
		return nil, err
	}

	return templates.GetQuery(res.Query), nil
}

func (ui *UI) NotFound(ctx context.Context, r *web.Request) (web.Template, error) {
	return templates.NotFound(), nil
}

func (ui *UI) ErrorHandler(ctx context.Context, r *web.Request, err error) web.Template {
	r.Log().Error("an unexpected error occurred", slog.String("error", err.Error()))

	return templates.Error()
}
