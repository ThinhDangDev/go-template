package transport

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ThinhDangDev/go-template/internal/boilerplate/auth"
	"github.com/ThinhDangDev/go-template/internal/boilerplate/store"

	"github.com/gin-gonic/gin"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *Server) healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":      "ok",
		"service":     s.runtime.Config.ServiceName,
		"environment": s.runtime.Config.Environment,
		"version":     s.runtime.Config.Version,
		"timestamp":   time.Now().UTC(),
	})
}

func (s *Server) readyz(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	if err := s.runtime.DB.PingContext(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not_ready",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

func (s *Server) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)
	if req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and password are required"})
		return
	}

	user, err := s.runtime.Users.GetByEmail(c.Request.Context(), req.Email)
	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		s.runtime.Logger.Error("failed to load user", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "user is inactive"})
		return
	}

	if err := auth.ComparePassword(user.PasswordHash, req.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := s.runtime.Tokens.IssueAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		s.runtime.Logger.Error("failed to issue jwt", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   int(s.runtime.Config.JWTAccessTTL.Seconds()),
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

func (s *Server) me(c *gin.Context) {
	user, ok := getCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing current user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"email":      user.Email,
		"role":       user.Role,
		"is_active":  user.IsActive,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	})
}

func (s *Server) rolePing(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, _ := getClaims(c)
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("%s route is accessible", requiredRole),
			"user": gin.H{
				"id":    claims.UserID,
				"email": claims.Email,
				"role":  claims.Role,
			},
		})
	}
}

func extractBearerToken(header string) (string, error) {
	parts := strings.Fields(strings.TrimSpace(header))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", fmt.Errorf("missing bearer token")
	}

	return parts[1], nil
}
