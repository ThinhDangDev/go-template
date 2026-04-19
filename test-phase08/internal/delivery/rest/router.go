package rest

import (
	"github.com/user/test-phase08/internal/delivery/rest/handler"
	"github.com/user/test-phase08/internal/delivery/rest/middleware"
	authMiddleware "github.com/user/test-phase08/internal/delivery/rest/middleware/auth"
	"github.com/user/test-phase08/internal/infrastructure/auth"
	"github.com/user/test-phase08/internal/infrastructure/logger"
	"github.com/user/test-phase08/internal/usecase"
	"github.com/gin-gonic/gin"
)

// NewRouter creates a new HTTP router
func NewRouter(
	userUsecase *usecase.UserUsecase,
	authUsecase *usecase.AuthUsecase,
	jwtService *auth.JWTService,
	logger logger.Logger,
) *gin.Engine {
	router := gin.New()

	// Middleware
	router.Use(middleware.Logger(logger))
	router.Use(middleware.Recovery(logger))
	router.Use(middleware.CORS())

	// Health check
	healthHandler := handler.NewHealthHandler()
	router.GET("/health", healthHandler.Check)

	// Metrics endpoint for Prometheus
	router.GET("/metrics", handler.MetricsHandler())

	// API v1
	v1 := router.Group("/api/v1")
	{
		// Auth routes (public)
		authHandler := handler.NewAuthHandler(authUsecase)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}
		// Protected routes requiring JWT
		protected := v1.Group("")
		protected.Use(authMiddleware.JWTMiddleware(jwtService))
		{
			protected.GET("/auth/me", authHandler.Me)

			// User routes (protected)
			userHandler := handler.NewUserHandler(userUsecase)
			users := protected.Group("/users")
			{
				users.POST("", userHandler.Create)
				users.GET("/:id", userHandler.GetByID)
				users.GET("", userHandler.List)
				users.PUT("/:id", userHandler.Update)
				users.DELETE("/:id", userHandler.Delete)
			}
		}
	}

	return router
}
