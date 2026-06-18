package main

import (
	"log"

	"github.com/gupta/leetcode-judge/internal/common"
	"github.com/gupta/leetcode-judge/internal/config"
	"github.com/gupta/leetcode-judge/internal/database"
	"github.com/gupta/leetcode-judge/internal/server"
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

	router := server.NewServer(cfg)
	server.Run(router, cfg)

}
