package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gupta/leetcode-judge/internal/auth"
	"github.com/gupta/leetcode-judge/internal/common"
	"github.com/gupta/leetcode-judge/internal/config"
)

func AuthRequired(jwtCfg *config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			common.Error(c, http.StatusUnauthorized, "missing authorization header", "UNAUTHORIZED")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			common.Error(c, http.StatusUnauthorized, "invalid authorization format", "UNAUTHORIZED")
			c.Abort()
			return
		}

		claims, err := auth.ValidateToken(parts[1], jwtCfg)
		if err != nil {
			common.Error(c, http.StatusUnauthorized, "invalid or expired token", "UNAUTHORIZED")
			c.Abort()
			return
		}

		// Store user info in context — any handler can access these
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}
