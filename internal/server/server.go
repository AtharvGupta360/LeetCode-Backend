package server

import (
    // Group 1: stdlib
    "context"
    "errors"
    "fmt"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    // Group 2: third-party
    "github.com/gin-gonic/gin"
    "github.com/jmoiron/sqlx"

    // Group 3: internal
    "github.com/gupta/leetcode-judge/internal/auth"
    "github.com/gupta/leetcode-judge/internal/common"
    "github.com/gupta/leetcode-judge/internal/config"
    "github.com/gupta/leetcode-judge/internal/handlers"
    "github.com/gupta/leetcode-judge/internal/middleware"
    "github.com/gupta/leetcode-judge/internal/repository"
    "github.com/gupta/leetcode-judge/internal/service"
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

	problemRepo    := repository.NewProblemRepository(db)
	problemService := service.NewProblemService(problemRepo)
	problemHandler := handlers.NewProblemHandler(problemService)

	// Public routes (no auth required)
	api := router.Group("/api/v1")
	{
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)
	}

	// Protected routes — any authenticated user
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

		// Problem read routes — any logged-in user
		problems := protected.Group("/problems")
		{
			problems.GET("", problemHandler.List)
			problems.GET("/:id", problemHandler.GetByID)
		}

		// Problem write routes — admin only
		adminProblems := protected.Group("/problems")
		adminProblems.Use(middleware.AdminRequired())
		{
			adminProblems.POST("", problemHandler.Create)
			adminProblems.PUT("/:id", problemHandler.Update)
			adminProblems.DELETE("/:id", problemHandler.Delete)
		}
	}

	return router
}

func Run(router *gin.Engine, cfg *config.Config) error {
    addr := fmt.Sprintf(":%d", cfg.Server.Port)

    // 1. Build our own http.Server instead of using router.Run()
    //    This gives us full control over its lifecycle.
    srv := &http.Server{
        Addr:         addr,
        Handler:      router,
        ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
        WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
        IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
    }

    // 2. Start the server in a goroutine so it doesn't block.
    //    We send any startup errors into a channel.
    serverErr := make(chan error, 1)
    go func() {
        common.Logger.Infof("server listening on %s", addr)
        if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
            serverErr <- err
        }
    }()

    // 3. Create a channel to receive OS signals.
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    // 4. Block here — wait for either a fatal server error OR a shutdown signal.
    select {
    case err := <-serverErr:
        return fmt.Errorf("server error: %w", err)
    case sig := <-quit:
        common.Logger.Infof("shutdown signal received: %s", sig)
    }

    // 5. We got a signal. Give in-flight requests time to finish.
    common.Logger.Info("gracefully shutting down server...")
    ctx, cancel := context.WithTimeout(
        context.Background(),
        time.Duration(cfg.Server.ShutdownTimeout)*time.Second,
    )
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        return fmt.Errorf("server forced to shutdown: %w", err)
    }

    common.Logger.Info("server stopped cleanly")
    return nil
}

