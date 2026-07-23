package handler

import (
	"net/http"

	"github.com/irvanmhndra/siaga-api/internal/service"
	"github.com/irvanmhndra/siaga-api/pkg/httputil"
	"github.com/labstack/echo/v5"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler { return &AuthHandler{svc} }

func (h *AuthHandler) Register(c *echo.Context) error {
	var req service.RegisterRequest
	if ok, err := bindValidate(c, &req); !ok {
		return err
	}
	res, err := h.svc.Register(c.Request().Context(), req)
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusCreated, res)
}

func (h *AuthHandler) Login(c *echo.Context) error {
	var req service.LoginRequest
	if ok, err := bindValidate(c, &req); !ok {
		return err
	}
	res, err := h.svc.Login(c.Request().Context(), req)
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusOK, res)
}

func (h *AuthHandler) Refresh(c *echo.Context) error {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = c.Bind(&body)
	res, err := h.svc.Refresh(c.Request().Context(), body.RefreshToken)
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusOK, res)
}

func (h *AuthHandler) Logout(c *echo.Context) error {
	_ = h.svc.Logout(c.Request().Context(), extractToken(c))
	return httputil.SuccessMsg(c, http.StatusOK, "logged out")
}

func (h *AuthHandler) Me(c *echo.Context) error {
	user, err := h.svc.Me(c.Request().Context(), getUserID(c))
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusOK, user)
}
