package main

import (
	"context"
	"log"

	"github.com/gupta/leetcode-judge/internal/cache"
	"github.com/gupta/leetcode-judge/internal/common"
	"github.com/gupta/leetcode-judge/internal/config"
	"github.com/gupta/leetcode-judge/internal/database"
	"github.com/gupta/leetcode-judge/internal/queue"
	"github.com/gupta/leetcode-judge/internal/repository"
	"github.com/gupta/leetcode-judge/internal/server"
	"github.com/gupta/leetcode-judge/internal/service/judge"
)

func main() {

	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("config load error: %v", err)
	}

	// fmt.Printf("Server will run on port %d", cfg.Server.Port)
	// fmt.Printf("Database: %s", cfg.DataBase.DSN())
	common.InitLogger(cfg.Server.Mode)
	defer common.Logger.Sync()
	common.Logger.Infof("server starting at port %d", cfg.Server.Port)
	common.Logger.Infof("database: %s", cfg.DataBase.DSN())
	common.Logger.Info("logger initialised successfully")
	db, err := database.NewPostgresConnection(&cfg.DataBase)
	if err != nil {
		common.Logger.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()
	common.Logger.Info("database connection established")

	if cfg.DataBase.AutoMigrate {
		if err := database.RunMigrations(db, "internal/database/migrations"); err != nil {
			common.Logger.Fatalf("migration failed: %v", err)
		}
	} else {
		common.Logger.Info("auto-migrate disabled, skipping migrations")
	}

	// Connect to Redis
	redisClient, err := cache.NewRedisClient(&cfg.Redis)
	if err != nil {
		common.Logger.Fatalf("redis connection failed: %v", err)
	}
	defer redisClient.Close()

	// Start judge worker(s) in background goroutines
	// They share a cancellable context so we can shut them down gracefully.
	judgeCtx, judgeCancel := context.WithCancel(context.Background())
	defer judgeCancel()

	runner, err := judge.NewRunner(cfg.Judge.TimeoutSeconds, int64(cfg.Judge.MemoryLimitMB))
	if err != nil {
		common.Logger.Fatalf("judge runner init failed: %v", err)
	}

	subRepo      := repository.NewSubmissionRepository(db)
	testCaseRepo := repository.NewTestCaseRepository(db)
	judgeQueue   := queue.NewRedisQueue(redisClient)
	judgeSvc      := judge.NewService(runner, subRepo, testCaseRepo, judgeQueue)

	workers := cfg.Judge.Workers
	if workers < 1 {
		workers = 1
	}
	for i := 0; i < workers; i++ {
		go judgeSvc.Start(judgeCtx)
	}
	common.Logger.Infof("started %d judge worker(s)", workers)

	router := server.NewServer(cfg, db, redisClient)
	if err := server.Run(router, cfg); err != nil {
    common.Logger.Errorf("server exited with error: %v", err)
}

	// Server has stopped — cancel judge workers and let them drain
	judgeCancel()
	common.Logger.Info("judge workers stopped")

}

