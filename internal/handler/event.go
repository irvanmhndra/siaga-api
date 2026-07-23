package handler

import (
	"net/http"

	"github.com/irvanmhndra/siaga-api/internal/service"
	"github.com/irvanmhndra/siaga-api/pkg/httputil"
	"github.com/labstack/echo/v5"
)

// EventHandler is the public Events API. It authenticates by routing_key
// (integration key) inside the body rather than a bearer token, so monitoring
// tools can post directly.
type EventHandler struct {
	svc *service.IncidentService
}

func NewEventHandler(svc *service.IncidentService) *EventHandler { return &EventHandler{svc} }

func (h *EventHandler) Enqueue(c *echo.Context) error {
	var req service.EventRequest
	if ok, err := bindValidate(c, &req); !ok {
		return err
	}
	res, err := h.svc.ProcessEvent(c.Request().Context(), req)
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusAccepted, res)
}
