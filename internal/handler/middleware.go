package handler

import (
	"net/http"
	"strings"

	"github.com/irvanmhndra/siaga-api/internal/service"
	"github.com/irvanmhndra/siaga-api/pkg/apperror"
	"github.com/irvanmhndra/siaga-api/pkg/httputil"
	"github.com/labstack/echo/v5"
)

const (
	userIDKey = "userID"
	orgIDKey  = "orgID"
)

func AuthMiddleware(authSvc *service.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			token := extractToken(c)
			if token == "" {
				return httputil.Error(c, http.StatusUnauthorized, "authentication required", "UNAUTHORIZED")
			}
			user, err := authSvc.ValidateToken(c.Request().Context(), token)
			if err != nil {
				if ae, ok := err.(*apperror.AppError); ok {
					return httputil.Error(c, ae.Code, ae.Message, ae.ErrCode)
				}
				return httputil.Error(c, http.StatusUnauthorized, "invalid token", "UNAUTHORIZED")
			}
			if user == nil {
				return httputil.Error(c, http.StatusUnauthorized, "invalid token", "UNAUTHORIZED")
			}
			c.Set(userIDKey, user.ID)
			c.Set(orgIDKey, user.OrganizationID)
			return next(c)
		}
	}
}

func getUserID(c *echo.Context) string {
	id, _ := c.Get(userIDKey).(string)
	return id
}

func getOrgID(c *echo.Context) string {
	id, _ := c.Get(orgIDKey).(string)
	return id
}

func extractToken(c *echo.Context) string {
	auth := c.Request().Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}
