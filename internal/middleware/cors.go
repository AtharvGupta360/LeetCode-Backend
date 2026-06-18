package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gupta/leetcode-judge/internal/common"
	"github.com/gupta/leetcode-judge/internal/config"
)

func CORS(cfg *config.CORSConfig) gin.HandlerFunc{
	allowOriginMap := make(map[string]bool,len(cfg.AllowedOrigins))
	allowAll := false 
	for _,origin:= range cfg.AllowedOrigins{
		if origin == "*"{
			allowAll = true 
			break
		}
		allowOriginMap[strings.ToLower(origin)] = true
	}
	methodStr := strings.Join(cfg.AllowedMethods,",")
	headerStr := strings.Join(cfg.AllowedHeaders,",")
	exposeStr := strings.Join(cfg.ExposedHeaders,",")
	
	return func(c *gin.Context){
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return 
		}
		originAllowed := allowAll || allowOriginMap[strings.ToLower(origin)]
		if !originAllowed {
			common.Logger.Warnf("CORS: rejected origin %s", origin)
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		// These headers go on EVERY cross-origin response, not just preflight
		c.Header("Access-Control-Allow-Origin", origin)

		if cfg.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if exposeStr != "" {
			c.Header("Access-Control-Expose-Headers", exposeStr)
		}

		// Handle preflight (OPTIONS) requests
		if c.Request.Method == http.MethodOptions {
			c.Header("Access-Control-Allow-Methods", methodStr)
			c.Header("Access-Control-Allow-Headers", headerStr)

			if cfg.MaxAge > 0 {
				c.Header("Access-Control-Max-Age", strconv.Itoa(cfg.MaxAge))
			}

			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
		
	}
	
}