package handler

import (
	"net/http"

	"github.com/irvanmhndra/siaga-api/internal/service"
	"github.com/irvanmhndra/siaga-api/pkg/httputil"
	"github.com/labstack/echo/v5"
)

type EscalationHandler struct {
	svc *service.EscalationService
}

func NewEscalationHandler(svc *service.EscalationService) *EscalationHandler {
	return &EscalationHandler{svc}
}

func (h *EscalationHandler) Create(c *echo.Context) error {
	var req service.PolicyRequest
	if ok, err := bindValidate(c, &req); !ok {
		return err
	}
	res, err := h.svc.Create(c.Request().Context(), getOrgID(c), req)
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusCreated, res)
}

func (h *EscalationHandler) List(c *echo.Context) error {
	res, err := h.svc.List(c.Request().Context(), getOrgID(c))
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusOK, res)
}

func (h *EscalationHandler) Get(c *echo.Context) error {
	res, err := h.svc.Get(c.Request().Context(), c.Param("id"), getOrgID(c))
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusOK, res)
}

func (h *EscalationHandler) Update(c *echo.Context) error {
	var req service.PolicyRequest
	if ok, err := bindValidate(c, &req); !ok {
		return err
	}
	res, err := h.svc.Update(c.Request().Context(), c.Param("id"), getOrgID(c), req)
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusOK, res)
}

func (h *EscalationHandler) Delete(c *echo.Context) error {
	if err := h.svc.Delete(c.Request().Context(), c.Param("id"), getOrgID(c)); err != nil {
		return respondErr(c, err)
	}
	return httputil.SuccessMsg(c, http.StatusOK, "escalation policy deleted")
}
