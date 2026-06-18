package middleware 

import(
	"time"
	"github.com/gin-gonic/gin"
	"github.com/gupta/leetcode-judge/internal/common"
)

func RequestLogger() gin.HandlerFunc{
	return func(c *gin.Context) {
		start := time.Now() 
		path := c.Request.URL.Path
		c.Next()
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		requestID, _ := c.Get(RequestIDHeader)
		common.Logger.Infof("[%s] %s %s → %d (%v)",
			requestID,
			c.Request.Method,
			path,
			statusCode,
			latency,
		)

	}

}