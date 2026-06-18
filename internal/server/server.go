package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gupta/leetcode-judge/internal/auth"
	"github.com/gupta/leetcode-judge/internal/common"
	"github.com/gupta/leetcode-judge/internal/config"
	"github.com/gupta/leetcode-judge/internal/handlers"
	"github.com/gupta/leetcode-judge/internal/middleware"
	"github.com/gupta/leetcode-judge/internal/repository"
	"github.com/jmoiron/sqlx"
)

func NewServer(cfg *config.Config, db *sqlx.DB) *gin.Engine {
	gin.SetMode(cfg.Server.Mode)
	router := gin.New()
	router.Use(middleware.CORS(&cfg.CORS))
	router.Use(middleware.RequestID())
	router.Use(middleware.RateLimiter(middleware.RateLimiterConfig{
		RequestsPerSecond: 10,
		BurstSize:         20,
	}))
	router.Use(middleware.Recovery())
	router.Use(middleware.RequestLogger())
	router.GET("/health", func(c *gin.Context) {
		common.Success(c, http.StatusOK, "server is healthy", gin.H{
			"status": "up",
		})
	})

	// Initialize dependencies
	userRepo := repository.NewUserRepository(db)
	authService := auth.NewService(userRepo, &cfg.JWT)
	authHandler := handlers.NewAuthHandler(authService)

	// Public routes (no auth required)
	api := router.Group("/api/v1")
	{
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)
	}

	// Protected routes (auth required)
	protected := api.Group("/")
	protected.Use(middleware.AuthRequired(&cfg.JWT))
	{
		protected.GET("/me", func(c *gin.Context) {
			common.Success(c, http.StatusOK, "authenticated user", gin.H{
				"userID":   c.GetString("userID"),
				"username": c.GetString("username"),
				"role":     c.GetString("role"),
			})
		})
	}

	return router
}

func Run(router *gin.Engine, cfg *config.Config) {
	addr := fmt.Sprintf(":%d", cfg.Server.Port)

	common.Logger.Infof("Server is running on %s ", addr)
	if err := router.Run(addr); err != nil {
		common.Logger.Fatalf("failed to start server %v", err)

	}

}
