package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/satria/obrolan-api/internal/config"
	"github.com/satria/obrolan-api/internal/utils"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.ErrorJSON(c, http.StatusUnauthorized, "missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.ErrorJSON(c, http.StatusUnauthorized, "invalid authorization format, expected: Bearer <token>")
			c.Abort()
			return
		}

		claims, err := utils.ValidateToken(cfg, parts[1])
		if err != nil {
			utils.ErrorJSON(c, http.StatusUnauthorized, "invalid or expired token")
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)

		c.Next()
	}
}

// GetUserID extracts user ID from gin context (set by AuthMiddleware)
func GetUserID(c *gin.Context) uuid.UUID {
	userID, _ := c.Get("userID")
	id, _ := userID.(uuid.UUID)
	return id
}

// GetUsername extracts username from gin context
func GetUsername(c *gin.Context) string {
	username, _ := c.Get("username")
	str, _ := username.(string)
	return str
}
