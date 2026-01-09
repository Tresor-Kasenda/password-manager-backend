package config

import (
	"log"
	"os"
	"strconv"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Email    EmailConfig
	HIBP     HIBPConfig
}

type ServerConfig struct {
	Port string
	Mode string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret     string
	ExpireTime int
}

type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type HIBPConfig struct {
	APIKey string
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("../config")

	viper.AutomaticEnv()

	viper.SetDefault("server.port", "8000")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("database.host", "127.0.0.1")
	viper.SetDefault("database.port", "5432")
	viper.SetDefault("database.dbname", "password_manager")
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", "6379")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Config file not found, using defaults and environment variables: %v", err)
	}

	config := &Config{
		Server: ServerConfig{
			Port: getEnvOrDefault("SERVER_PORT", viper.GetString("server.port")),
			Mode: getEnvOrDefault("SERVER_MODE", viper.GetString("server.mode")),
		},
		Database: DatabaseConfig{
			Host:     getEnvOrDefault("DATABASE_HOST", viper.GetString("database.host")),
			Port:     getEnvOrDefault("DATABASE_PORT", viper.GetString("database.port")),
			User:     getEnvOrDefault("DATABASE_USER", viper.GetString("database.user")),
			Password: getEnvOrDefault("DATABASE_PASSWORD", viper.GetString("database.password")),
			DBName:   getEnvOrDefault("DATABASE_DBNAME", viper.GetString("database.dbname")),
			SSLMode:  getEnvOrDefault("DATABASE_SSLMODE", viper.GetString("database.sslmode")),
		},
		Redis: RedisConfig{
			Host:     getEnvOrDefault("REDIS_HOST", viper.GetString("redis.host")),
			Port:     getEnvOrDefault("REDIS_PORT", viper.GetString("redis.port")),
			Password: getEnvOrDefault("REDIS_PASSWORD", viper.GetString("redis.password")),
			DB:       getEnvIntOrDefault("REDIS_DB", viper.GetInt("redis.db")),
		},
		JWT: JWTConfig{
			Secret:     getEnvOrDefault("JWT_SECRET", viper.GetString("jwt.secret")),
			ExpireTime: getEnvIntOrDefault("JWT_EXPIRE_TIME", viper.GetInt("jwt.expire_time")),
		},
		Email: EmailConfig{
			Host:     getEnvOrDefault("EMAIL_HOST", viper.GetString("email.host")),
			Port:     getEnvIntOrDefault("EMAIL_PORT", viper.GetInt("email.port")),
			Username: getEnvOrDefault("EMAIL_USERNAME", viper.GetString("email.username")),
			Password: getEnvOrDefault("EMAIL_PASSWORD", viper.GetString("email.password")),
			From:     getEnvOrDefault("EMAIL_FROM", viper.GetString("email.from")),
		},
		HIBP: HIBPConfig{
			APIKey: getEnvOrDefault("HIBP_API_KEY", viper.GetString("hibp.api_key")),
		},
	}

	if config.Database.DBName == "" {
		config.Database.DBName = viper.GetString("database.dbname")
	}

	return config, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
