package database

import (
	"fmt"
	"log"
	"net/url"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"github.com/tresor/password-manager/internal/config"
	"github.com/tresor/password-manager/internal/models"
)

func NewPostgresDB(cfg *config.DatabaseConfig) (*sqlx.DB, error) {
	u := &url.URL{
		Scheme: "postgres",
		Host:   fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Path:   "/" + cfg.DBName,
	}
	if cfg.Password != "" {
		u.User = url.UserPassword(cfg.User, cfg.Password)
	} else {
		u.User = url.User(cfg.User)
	}
	q := u.Query()
	if cfg.SSLMode != "" {
		q.Set("sslmode", cfg.SSLMode)
	}
	u.RawQuery = q.Encode()

	dsn := u.String()
	log.Printf("sqlx DSN: %s", dsn)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection established")
	return db, nil
}

// NewGormDB creates a new GORM database connection with auto-migration
func NewGormDB(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	// Log config values for debugging
	log.Printf("Database Config - Host: %s, Port: %s, User: %s, DBName: [%s], SSLMode: %s",
		cfg.Host, cfg.Port, cfg.User, cfg.DBName, cfg.SSLMode)

	// Build DSN for GORM using proper format
	var dsn string
	if cfg.Password != "" {
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
			cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode)
	} else {
		dsn = fmt.Sprintf("host=%s user=%s dbname=%s port=%s sslmode=%s",
			cfg.Host, cfg.User, cfg.DBName, cfg.Port, cfg.SSLMode)
	}
	log.Printf("Built DSN: %s", dsn)

	// Open GORM connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database with GORM: %w", err)
	}

	log.Println("GORM database connection established")

	// Auto-migrate tables
	if err := db.AutoMigrate(
		&models.User{},
		&models.Vault{},
		&models.SharedPassword{},
	); err != nil {
		return nil, fmt.Errorf("failed to run auto-migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)

	return db, nil
}
