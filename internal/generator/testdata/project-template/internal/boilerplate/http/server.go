package transport

import (
	"context"
	"net/http"
	"time"

	"__MODULE_PATH__/internal/api/handler"
	"__MODULE_PATH__/internal/boilerplate/app"
	pb "__MODULE_PATH__/protogen"

	"github.com/gin-gonic/gin"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

type Server struct {
	runtime         *app.Runtime
	templateService pb.TemplateServiceServer
	gatewayHandler  http.Handler
}

func NewServer(runtime *app.Runtime) *Server {
	return &Server{
		runtime:         runtime,
		templateService: handler.NewTemplateService(runtime.AuthUseCase, runtime.AdminUseCase, runtime.SystemUseCase),
	}
}

func (s *Server) Handler() (*gin.Engine, error) {
	if s.runtime.Config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	gatewayHandler, err := s.newGatewayHandler()
	if err != nil {
		return nil, err
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

	router.GET("/swagger.json", func(c *gin.Context) {
		c.File(s.runtime.Config.ResolvedSwaggerJSONPath())
	})
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/healthz", s.healthz)
	router.GET("/readyz", s.readyz)

	router.GET("/api/v1/public/ping", gin.WrapH(gatewayHandler))
	router.POST("/api/v1/auth/register", gin.WrapH(gatewayHandler))
	router.POST("/api/v1/auth/login", gin.WrapH(gatewayHandler))

	protected := router.Group("/api/v1")
	protected.Use(s.Authenticate(), s.RequireRBAC())
	protected.GET("/auth/me", gin.WrapH(gatewayHandler))
	protected.GET("/admin/ping", gin.WrapH(gatewayHandler))
	protected.GET("/admin/users", gin.WrapH(gatewayHandler))
	protected.GET("/admin/roles", gin.WrapH(gatewayHandler))
	protected.PATCH("/admin/users/:user_id/access", gin.WrapH(gatewayHandler))

	return router, nil
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

func (s *Server) GRPCServer() *grpc.Server {
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(s.AuthUnaryInterceptor()),
	)
	pb.RegisterTemplateServiceServer(server, s.templateService)
	return server
}

func (s *Server) newGatewayHandler() (http.Handler, error) {
	if s.gatewayHandler != nil {
		return s.gatewayHandler, nil
	}

	mux := gwruntime.NewServeMux()
	if err := pb.RegisterTemplateServiceHandlerServer(context.Background(), mux, s.templateService); err != nil {
		return nil, err
	}

	s.gatewayHandler = mux
	return s.gatewayHandler, nil
}
