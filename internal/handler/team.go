package handler

import (
	"net/http"

	"github.com/irvanmhndra/siaga-api/internal/service"
	"github.com/irvanmhndra/siaga-api/pkg/httputil"
	"github.com/labstack/echo/v5"
)

type TeamHandler struct {
	svc *service.TeamService
}

func NewTeamHandler(svc *service.TeamService) *TeamHandler { return &TeamHandler{svc} }

func (h *TeamHandler) Create(c *echo.Context) error {
	var req service.TeamRequest
	if ok, err := bindValidate(c, &req); !ok {
		return err
	}
	res, err := h.svc.Create(c.Request().Context(), getOrgID(c), req)
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusCreated, res)
}

func (h *TeamHandler) List(c *echo.Context) error {
	res, err := h.svc.List(c.Request().Context(), getOrgID(c))
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusOK, res)
}

func (h *TeamHandler) Get(c *echo.Context) error {
	res, err := h.svc.Get(c.Request().Context(), c.Param("id"), getOrgID(c))
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusOK, res)
}

func (h *TeamHandler) Update(c *echo.Context) error {
	var req service.TeamRequest
	if ok, err := bindValidate(c, &req); !ok {
		return err
	}
	res, err := h.svc.Update(c.Request().Context(), c.Param("id"), getOrgID(c), req)
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusOK, res)
}

func (h *TeamHandler) Delete(c *echo.Context) error {
	if err := h.svc.Delete(c.Request().Context(), c.Param("id"), getOrgID(c)); err != nil {
		return respondErr(c, err)
	}
	return httputil.SuccessMsg(c, http.StatusOK, "team deleted")
}

func (h *TeamHandler) AddMember(c *echo.Context) error {
	var req struct {
		UserID string `json:"user_id" validate:"required"`
	}
	if ok, err := bindValidate(c, &req); !ok {
		return err
	}
	if err := h.svc.AddMember(c.Request().Context(), c.Param("id"), getOrgID(c), req.UserID); err != nil {
		return respondErr(c, err)
	}
	return httputil.SuccessMsg(c, http.StatusOK, "member added")
}

func (h *TeamHandler) RemoveMember(c *echo.Context) error {
	if err := h.svc.RemoveMember(c.Request().Context(), c.Param("id"), getOrgID(c), c.Param("userId")); err != nil {
		return respondErr(c, err)
	}
	return httputil.SuccessMsg(c, http.StatusOK, "member removed")
}
