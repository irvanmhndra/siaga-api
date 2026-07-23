package handler

import (
	"log/slog"
	"net/http"

	"github.com/irvanmhndra/siaga-api/pkg/apperror"
	"github.com/irvanmhndra/siaga-api/pkg/httputil"
	"github.com/labstack/echo/v5"
)

// respondErr maps a service error to the proper HTTP response.
func respondErr(c *echo.Context, err error) error {
	if ae, ok := err.(*apperror.AppError); ok {
		return httputil.Error(c, ae.Code, ae.Message, ae.ErrCode)
	}
	slog.Error("internal error", "path", c.Request().URL.Path, "error", err)
	return httputil.Error(c, http.StatusInternalServerError, "internal error", "INTERNAL_ERROR")
}

// bindValidate binds and validates the request body. When it returns ok=false
// it has already written the error response, so the handler should
// `return err` immediately.
func bindValidate(c *echo.Context, req interface{}) (ok bool, err error) {
	if err := c.Bind(req); err != nil {
		return false, httputil.Error(c, http.StatusBadRequest, "invalid request", "BIND_ERROR")
	}
	if err := c.Validate(req); err != nil {
		return false, httputil.Error(c, http.StatusUnprocessableEntity, err.Error(), "VALIDATION_ERROR")
	}
	return true, nil
}
