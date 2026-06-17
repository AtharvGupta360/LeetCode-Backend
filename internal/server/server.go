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
	corsConfig := middleware.CORSConfig{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   cfg.CORS.AllowedMethods,
		AllowedHeaders:   cfg.CORS.AllowedHeaders,
		ExposedHeaders:   cfg.CORS.ExposedHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           cfg.CORS.MaxAge,
	}
	router.Use(middleware.CORS(&corsConfig))

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
