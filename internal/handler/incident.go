package handler

import (
	"net/http"

	"github.com/irvanmhndra/siaga-api/internal/repository"
	"github.com/irvanmhndra/siaga-api/internal/service"
	"github.com/irvanmhndra/siaga-api/pkg/httputil"
	"github.com/labstack/echo/v5"
)

type IncidentHandler struct {
	svc *service.IncidentService
}

func NewIncidentHandler(svc *service.IncidentService) *IncidentHandler { return &IncidentHandler{svc} }

func (h *IncidentHandler) List(c *echo.Context) error {
	f := repository.IncidentFilter{
		Status:    c.QueryParam("status"),
		ServiceID: c.QueryParam("service_id"),
	}
	res, err := h.svc.List(c.Request().Context(), getOrgID(c), f)
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusOK, res)
}

func (h *IncidentHandler) Get(c *echo.Context) error {
	res, err := h.svc.Get(c.Request().Context(), c.Param("id"), getOrgID(c))
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusOK, res)
}

func (h *IncidentHandler) Acknowledge(c *echo.Context) error {
	res, err := h.svc.Acknowledge(c.Request().Context(), c.Param("id"), getOrgID(c), getUserID(c))
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusOK, res)
}

func (h *IncidentHandler) Resolve(c *echo.Context) error {
	res, err := h.svc.Resolve(c.Request().Context(), c.Param("id"), getOrgID(c), getUserID(c))
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusOK, res)
}
