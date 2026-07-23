package router

import (
	"net/http"

	"github.com/irvanmhndra/siaga-api/internal/handler"
	"github.com/irvanmhndra/siaga-api/internal/service"
	"github.com/labstack/echo/v5"
)

// Handlers bundles every HTTP handler so Register's signature stays stable as
// the surface grows.
type Handlers struct {
	Auth         *handler.AuthHandler
	User         *handler.UserHandler
	Team         *handler.TeamHandler
	Schedule     *handler.ScheduleHandler
	Escalation   *handler.EscalationHandler
	Service      *handler.ServiceHandler
	Incident     *handler.IncidentHandler
	Event        *handler.EventHandler
	Notification *handler.NotificationHandler
}

func Register(e *echo.Echo, authSvc *service.AuthService, h Handlers) {
	e.GET("/health", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	v1 := e.Group("/api/v1")
	mid := handler.AuthMiddleware(authSvc)

	// Auth
	auth := v1.Group("/auth")
	auth.POST("/register", h.Auth.Register)
	auth.POST("/login", h.Auth.Login)
	auth.POST("/refresh", h.Auth.Refresh)
	auth.POST("/logout", h.Auth.Logout, mid)
	v1.GET("/me", h.Auth.Me, mid)

	// Public Events API — authenticated by routing_key in the body.
	v1.POST("/events/enqueue", h.Event.Enqueue)

	// Everything below requires a bearer token.
	// Users / profile
	v1.GET("/users", h.User.List, mid)
	v1.POST("/users", h.User.Invite, mid)
	v1.PUT("/profile", h.User.UpdateProfile, mid)

	// Teams
	teams := v1.Group("/teams", mid)
	teams.POST("", h.Team.Create)
	teams.GET("", h.Team.List)
	teams.GET("/:id", h.Team.Get)
	teams.PUT("/:id", h.Team.Update)
	teams.DELETE("/:id", h.Team.Delete)
	teams.POST("/:id/members", h.Team.AddMember)
	teams.DELETE("/:id/members/:userId", h.Team.RemoveMember)

	// Schedules
	sched := v1.Group("/schedules", mid)
	sched.POST("", h.Schedule.Create)
	sched.GET("", h.Schedule.List)
	sched.GET("/:id", h.Schedule.Get)
	sched.PUT("/:id", h.Schedule.Update)
	sched.DELETE("/:id", h.Schedule.Delete)
	sched.GET("/:id/on-call", h.Schedule.OnCall)
	sched.POST("/:id/overrides", h.Schedule.AddOverride)
	sched.DELETE("/:id/overrides/:overrideId", h.Schedule.DeleteOverride)

	// Escalation policies
	esc := v1.Group("/escalation-policies", mid)
	esc.POST("", h.Escalation.Create)
	esc.GET("", h.Escalation.List)
	esc.GET("/:id", h.Escalation.Get)
	esc.PUT("/:id", h.Escalation.Update)
	esc.DELETE("/:id", h.Escalation.Delete)

	// Services
	svc := v1.Group("/services", mid)
	svc.POST("", h.Service.Create)
	svc.GET("", h.Service.List)
	svc.GET("/:id", h.Service.Get)
	svc.PUT("/:id", h.Service.Update)
	svc.DELETE("/:id", h.Service.Delete)

	// Incidents
	inc := v1.Group("/incidents", mid)
	inc.GET("", h.Incident.List)
	inc.GET("/:id", h.Incident.Get)
	inc.POST("/:id/acknowledge", h.Incident.Acknowledge)
	inc.POST("/:id/resolve", h.Incident.Resolve)

	// Notifications (in-app inbox)
	notif := v1.Group("/notifications", mid)
	notif.GET("", h.Notification.List)
	notif.POST("/read-all", h.Notification.MarkAllRead)
	notif.POST("/:id/read", h.Notification.MarkRead)
}
