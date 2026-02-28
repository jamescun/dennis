package templates

import (
	"embed"
	"net/http"
)

//go:embed css
var assets embed.FS

// Assets returns an HTTP Handler that serves static CSS and Images to support
// the applications templates. prefix will automatically be stripped from the
// request path before serving.
func Assets(prefix string) http.Handler {
	return http.StripPrefix(prefix, http.FileServerFS(assets))
}
