package database

import (
	"context"
	"fmt"
	"time"

	"github.com/gupta/leetcode-judge/internal/common"
	"github.com/gupta/leetcode-judge/internal/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)
func NewPostgresConnection(cfg *config.DatabaseConfig) (*sqlx.DB, error) {
	dsn := cfg.DSN()
	common.Logger.Infof("connecting to database: %s", cfg.Host)
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	common.Logger.Info(" Database connected successfully")
	return db, nil
}