package transport

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"__MODULE_PATH__/internal/boilerplate/auth"
	"__MODULE_PATH__/internal/boilerplate/store"
	"__MODULE_PATH__/internal/boilerplate/telemetry"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const (
	requestIDHeader  = "X-Request-ID"
	claimsContextKey = "auth_claims"
	userContextKey   = "auth_user"
)

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(requestIDHeader)
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Set(requestIDHeader, requestID)
		c.Writer.Header().Set(requestIDHeader, requestID)
		c.Next()
	}
}

func TracingMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID, _ := c.Get(requestIDHeader)
		requestIDStr, _ := requestID.(string)

		ctx, span := otel.Tracer(serviceName).Start(
			c.Request.Context(),
			fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path),
			oteltrace.WithAttributes(
				attribute.String("http.method", c.Request.Method),
				attribute.String("http.target", c.Request.URL.Path),
				attribute.String("request.id", requestIDStr),
			),
		)
		defer span.End()

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}

		status := c.Writer.Status()
		span.SetAttributes(
			attribute.String("http.route", route),
			attribute.Int("http.status_code", status),
		)
		if status >= http.StatusInternalServerError {
			span.SetStatus(codes.Error, http.StatusText(status))
		}
	}
}

func MetricsMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		done := telemetry.InFlightRequests(serviceName)
		defer done()

		start := time.Now()
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}

		telemetry.ObserveHTTPRequest(
			serviceName,
			c.Request.Method,
			route,
			c.Writer.Status(),
			time.Since(start),
		)
	}
}

func LoggingMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		requestID, _ := c.Get(requestIDHeader)
		requestIDStr, _ := requestID.(string)

		claims, _ := getClaims(c)
		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}

		traceID := oteltrace.SpanFromContext(c.Request.Context()).SpanContext().TraceID().String()
		log := logger.With(
			"request_id", requestIDStr,
			"trace_id", traceID,
			"method", c.Request.Method,
			"path", route,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
		)
		if claims != nil {
			log = log.With("user_id", claims.UserID, "role", claims.Role)
		}

		log.Info("http request completed")
	}
}

func RecoveryMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		logger.Error(
			"panic recovered",
			"panic", recovered,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
		)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	})
}

func (s *Server) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := extractBearerToken(c.GetHeader("Authorization"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		claims, err := s.runtime.Tokens.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		user, err := s.runtime.Users.GetByID(c.Request.Context(), claims.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}
		if !user.IsActive {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user is inactive"})
			return
		}

		claims.Email = user.Email
		claims.Role = user.Role

		c.Set(claimsContextKey, claims)
		c.Set(userContextKey, user)
		c.Next()
	}
}

func (s *Server) RequireRBAC() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := getClaims(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing auth claims"})
			return
		}

		allowed, err := s.runtime.Authorizer.Authorize(claims.Role, c.Request.URL.Path, c.Request.Method)
		if err != nil {
			s.runtime.Logger.Error("rbac authorization failed", "error", err)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		c.Next()
	}
}

func getClaims(c *gin.Context) (*auth.Claims, bool) {
	value, exists := c.Get(claimsContextKey)
	if !exists {
		return nil, false
	}

	claims, ok := value.(*auth.Claims)
	return claims, ok
}

func getCurrentUser(c *gin.Context) (*store.User, bool) {
	value, exists := c.Get(userContextKey)
	if !exists {
		return nil, false
	}

	user, ok := value.(*store.User)
	return user, ok
}
