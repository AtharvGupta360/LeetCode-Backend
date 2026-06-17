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
		
	
		common.Logger.Infof("%s %s → %d (%v)",
			c.Request.Method,
			path,
			statusCode,
			latency,
	)
	}

}