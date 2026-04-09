package database

import (
	"fmt"
	"log"

	"github.com/mortogo321/go-fiber-api/config"
	"github.com/mortogo321/go-fiber-api/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the global database connection pool.
var DB *gorm.DB

// InitDB opens a PostgreSQL connection and runs auto-migrations.
func InitDB(cfg *config.Config) error {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Run auto-migrations for all models.
	if err := db.AutoMigrate(&models.User{}, &models.Product{}); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	DB = db
	log.Println("Database connected and migrated successfully")
	return nil
}
