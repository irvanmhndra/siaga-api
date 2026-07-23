package handler

import (
	"net/http"

	"github.com/irvanmhndra/siaga-api/internal/service"
	"github.com/irvanmhndra/siaga-api/pkg/httputil"
	"github.com/labstack/echo/v5"
)

type ScheduleHandler struct {
	svc *service.ScheduleService
}

func NewScheduleHandler(svc *service.ScheduleService) *ScheduleHandler { return &ScheduleHandler{svc} }

func (h *ScheduleHandler) Create(c *echo.Context) error {
	var req service.ScheduleRequest
	if ok, err := bindValidate(c, &req); !ok {
		return err
	}
	res, err := h.svc.Create(c.Request().Context(), getOrgID(c), req)
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusCreated, res)
}

func (h *ScheduleHandler) List(c *echo.Context) error {
	res, err := h.svc.List(c.Request().Context(), getOrgID(c))
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusOK, res)
}

func (h *ScheduleHandler) Get(c *echo.Context) error {
	res, err := h.svc.Get(c.Request().Context(), c.Param("id"), getOrgID(c))
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusOK, res)
}

func (h *ScheduleHandler) Update(c *echo.Context) error {
	var req service.ScheduleRequest
	if ok, err := bindValidate(c, &req); !ok {
		return err
	}
	res, err := h.svc.Update(c.Request().Context(), c.Param("id"), getOrgID(c), req)
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusOK, res)
}

func (h *ScheduleHandler) Delete(c *echo.Context) error {
	if err := h.svc.Delete(c.Request().Context(), c.Param("id"), getOrgID(c)); err != nil {
		return respondErr(c, err)
	}
	return httputil.SuccessMsg(c, http.StatusOK, "schedule deleted")
}

func (h *ScheduleHandler) OnCall(c *echo.Context) error {
	userID, err := h.svc.OnCallNow(c.Request().Context(), c.Param("id"), getOrgID(c))
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusOK, map[string]string{"on_call_user_id": userID})
}

func (h *ScheduleHandler) AddOverride(c *echo.Context) error {
	var req service.OverrideRequest
	if ok, err := bindValidate(c, &req); !ok {
		return err
	}
	res, err := h.svc.AddOverride(c.Request().Context(), c.Param("id"), getOrgID(c), req)
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusCreated, res)
}

func (h *ScheduleHandler) DeleteOverride(c *echo.Context) error {
	err := h.svc.DeleteOverride(c.Request().Context(), c.Param("id"), getOrgID(c), c.Param("overrideId"))
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.SuccessMsg(c, http.StatusOK, "override deleted")
}
