package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gupta/leetcode-judge/internal/common"
	"github.com/gupta/leetcode-judge/internal/config"
	"github.com/gupta/leetcode-judge/internal/middleware"
)

func NewServer(cfg *config.Config) *gin.Engine {
	gin.SetMode(cfg.Server.Mode)
	router := gin.New()
	router.Use(middleware.CORS(&cfg.CORS))
	router.Use(middleware.RequestID())
	router.Use(middleware.Recovery())
	router.Use(middleware.RequestLogger())
	router.GET("/health", func(c *gin.Context) {
		common.Success(c, http.StatusOK, "server is healthy", gin.H{
			"status": "up",
		})
	})
	return router
}

func Run(router *gin.Engine, cfg *config.Config) {
	addr := fmt.Sprintf(":%d", cfg.Server.Port)

	common.Logger.Infof("Server is running on %s ", addr)
	if err := router.Run(addr); err != nil {
		common.Logger.Fatalf("failed to start server %v", err)

	}

}
