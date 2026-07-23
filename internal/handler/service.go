package handler

import (
	"net/http"

	"github.com/irvanmhndra/siaga-api/internal/service"
	"github.com/irvanmhndra/siaga-api/pkg/httputil"
	"github.com/labstack/echo/v5"
)

type ServiceHandler struct {
	svc *service.ServiceService
}

func NewServiceHandler(svc *service.ServiceService) *ServiceHandler { return &ServiceHandler{svc} }

func (h *ServiceHandler) Create(c *echo.Context) error {
	var req service.ServiceRequest
	if ok, err := bindValidate(c, &req); !ok {
		return err
	}
	res, err := h.svc.Create(c.Request().Context(), getOrgID(c), req)
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusCreated, res)
}

func (h *ServiceHandler) List(c *echo.Context) error {
	res, err := h.svc.List(c.Request().Context(), getOrgID(c))
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusOK, res)
}

func (h *ServiceHandler) Get(c *echo.Context) error {
	res, err := h.svc.Get(c.Request().Context(), c.Param("id"), getOrgID(c))
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusOK, res)
}

func (h *ServiceHandler) Update(c *echo.Context) error {
	var req service.ServiceRequest
	if ok, err := bindValidate(c, &req); !ok {
		return err
	}
	res, err := h.svc.Update(c.Request().Context(), c.Param("id"), getOrgID(c), req)
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusOK, res)
}

func (h *ServiceHandler) Delete(c *echo.Context) error {
	if err := h.svc.Delete(c.Request().Context(), c.Param("id"), getOrgID(c)); err != nil {
		return respondErr(c, err)
	}
	return httputil.SuccessMsg(c, http.StatusOK, "service deleted")
}
