package handler

import (
	"net/http"

	"github.com/irvanmhndra/siaga-api/internal/service"
	"github.com/irvanmhndra/siaga-api/pkg/httputil"
	"github.com/labstack/echo/v5"
)

type NotificationHandler struct {
	svc *service.NotificationService
}

func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc}
}

func (h *NotificationHandler) List(c *echo.Context) error {
	unreadOnly := c.QueryParam("unread") == "true"
	res, err := h.svc.List(c.Request().Context(), getUserID(c), unreadOnly)
	if err != nil {
		return respondErr(c, err)
	}
	count, err := h.svc.CountUnread(c.Request().Context(), getUserID(c))
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusOK, map[string]any{"items": res, "unread": count})
}

func (h *NotificationHandler) MarkRead(c *echo.Context) error {
	if err := h.svc.MarkRead(c.Request().Context(), c.Param("id"), getUserID(c)); err != nil {
		return respondErr(c, err)
	}
	return httputil.SuccessMsg(c, http.StatusOK, "marked as read")
}

func (h *NotificationHandler) MarkAllRead(c *echo.Context) error {
	if err := h.svc.MarkAllRead(c.Request().Context(), getUserID(c)); err != nil {
		return respondErr(c, err)
	}
	return httputil.SuccessMsg(c, http.StatusOK, "all marked as read")
}
