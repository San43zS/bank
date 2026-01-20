package middleware

import (
	"context"
	"net/http"

	"banking-platform/internal/service"
	"github.com/gin-gonic/gin"
)

type userIDContextKey struct{}

func AuthMiddleware(authService service.IAuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			respondWithError(c, "authorization header required", http.StatusUnauthorized)
			c.Abort()
			return
		}

		token := extractTokenFromHeader(authHeader)
		if token == "" {
			respondWithError(c, "invalid authorization header format", http.StatusUnauthorized)
			c.Abort()
			return
		}

		ctx := c.Request.Context()
		userID, err := authService.ValidateToken(ctx, token)
		if err != nil {
			respondWithError(c, "invalid or expired token", http.StatusUnauthorized)
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Request = c.Request.WithContext(context.WithValue(ctx, userIDContextKey{}, userID))
		c.Next()
	}
}

func extractTokenFromHeader(authHeader string) string {
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	return ""
}

func respondWithError(c *gin.Context, message string, statusCode int) {
	c.JSON(statusCode, gin.H{"error": message})
}
