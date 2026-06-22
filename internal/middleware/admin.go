package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gupta/leetcode-judge/internal/common"
)

// AdminRequired must be chained AFTER AuthRequired.
// It rejects requests where the authenticated user is not an admin.
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role != "admin" {
			common.Error(c, http.StatusForbidden, "admin access required", "FORBIDDEN")
			c.Abort()
			return
		}
		c.Next()
	}
}
