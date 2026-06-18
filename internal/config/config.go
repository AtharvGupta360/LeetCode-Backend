package config

import (
	"fmt"
	"log"

	"github.com/gupta/leetcode-judge/internal/middleware"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	DataBase DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	CORS     middleware.CORSConfig `mapstructure:"cors"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Host        string `mapstructure:"host"`
	DBName      string `mapstructure:"dbname"`
	SSLMode     string `mapstructure:"sslmode"`
	Port        int    `mapstructure:"port"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	AutoMigrate bool   `mapstructure:"autoMigrate"`
}


type JWTConfig struct {
	SecretKey   string `mapstructure:"secretkey"`
	ExpiryHours int    `mapstructure:"expiryHours"`
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)
	viper.SetDefault("cors.allowedOrigins", []string{"http://localhost:3000", "http://localhost:5173"})
	viper.SetDefault("cors.allowedMethods", []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"})
	viper.SetDefault("cors.allowedHeaders", []string{"Content-Type", "Authorization", "X-Request-ID"})
	viper.SetDefault("cors.exposedHeaders", []string{"X-Request-ID"})
	viper.SetDefault("cors.allowCredentials", true)
	viper.SetDefault("cors.maxAge", 86400)


	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.dbname", "leetcode_judge")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.autoMigrate", true)
	viper.SetDefault("jwt.secretkey", "dev-secret-change-in-prod")
	viper.SetDefault("jwt.expiryHours", 24)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config: %w", err)
		}
		log.Println("No config file found, using defaults + env vars")
	}
	viper.AutomaticEnv()
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("config unmarshal error : %w", err)
	}
	fmt.Println("config loaded successfully")
	return &cfg, nil

}

func(c *DatabaseConfig)DSN()string{

	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",c.Host,c.Port,c.User,c.Password,c.DBName,c.SSLMode)
}