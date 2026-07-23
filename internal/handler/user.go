package handler

import (
	"net/http"

	"github.com/irvanmhndra/siaga-api/internal/service"
	"github.com/irvanmhndra/siaga-api/pkg/httputil"
	"github.com/labstack/echo/v5"
)

type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler { return &UserHandler{svc} }

func (h *UserHandler) List(c *echo.Context) error {
	res, err := h.svc.ListByOrg(c.Request().Context(), getOrgID(c))
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusOK, res)
}

func (h *UserHandler) Invite(c *echo.Context) error {
	var req service.InviteRequest
	if ok, err := bindValidate(c, &req); !ok {
		return err
	}
	res, err := h.svc.Invite(c.Request().Context(), getOrgID(c), req)
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusCreated, res)
}

func (h *UserHandler) UpdateProfile(c *echo.Context) error {
	var req service.ProfileRequest
	if ok, err := bindValidate(c, &req); !ok {
		return err
	}
	res, err := h.svc.UpdateProfile(c.Request().Context(), getUserID(c), req)
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusOK, res)
}
