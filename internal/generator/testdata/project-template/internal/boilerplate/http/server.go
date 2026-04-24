package transport

import (
	"net/http"
	"time"

	"__MODULE_PATH__/internal/boilerplate/app"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	runtime *app.Runtime
}

func NewServer(runtime *app.Runtime) *Server {
	return &Server{runtime: runtime}
}

func (s *Server) Handler() *gin.Engine {
	if s.runtime.Config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(
		RecoveryMiddleware(s.runtime.Logger),
		RequestIDMiddleware(),
		TracingMiddleware(s.runtime.Config.ServiceName),
		MetricsMiddleware(s.runtime.Config.ServiceName),
		LoggingMiddleware(s.runtime.Logger),
	)

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
	})

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/healthz", s.healthz)
	router.GET("/readyz", s.readyz)

	api := router.Group("/api/v1")
	api.GET("/public/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "public pong"})
	})

	authGroup := api.Group("/auth")
	authGroup.POST("/login", s.login)
	authGroup.GET("/me", s.Authenticate(), s.RequireRBAC(), s.me)

	adminGroup := api.Group("/admin", s.Authenticate(), s.RequireRBAC())
	adminGroup.GET("/ping", s.rolePing("admin"))

	operatorGroup := api.Group("/operator", s.Authenticate(), s.RequireRBAC())
	operatorGroup.GET("/ping", s.rolePing("operator"))

	viewerGroup := api.Group("/viewer", s.Authenticate(), s.RequireRBAC())
	viewerGroup.GET("/ping", s.rolePing("viewer"))

	return router
}

func (s *Server) HTTPServer(handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              s.runtime.Config.HTTPAddress(),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       s.runtime.Config.HTTPReadTimeout,
		WriteTimeout:      s.runtime.Config.HTTPWriteTimeout,
		IdleTimeout:       60 * time.Second,
	}
}
