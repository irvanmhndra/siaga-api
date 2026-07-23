package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v5"
)

// SPAHandler serves the built single-page app. Requests that map to a real file
// in the dist directory are served directly; everything else falls back to
// index.html so client-side routing works on hard refresh.
//
// It uses http.ServeFile with absolute OS paths rather than echo's c.File,
// which resolves through an fs.FS that rejects absolute paths.
type SPAHandler struct {
	dist  string
	index string
}

func NewSPAHandler(dist string) *SPAHandler {
	clean := filepath.Clean(dist)
	return &SPAHandler{dist: clean, index: filepath.Join(clean, "index.html")}
}

func (h *SPAHandler) Serve(c *echo.Context) error {
	full := filepath.Join(h.dist, filepath.Clean("/"+c.Request().URL.Path))
	// Fall back to the SPA shell for traversal attempts, missing files, and dirs.
	if !strings.HasPrefix(full, h.dist) {
		full = h.index
	} else if info, err := os.Stat(full); err != nil || info.IsDir() {
		full = h.index
	}
	http.ServeFile(c.Response(), c.Request(), full)
	return nil
}
