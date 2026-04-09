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

// InitDB opens a GORM connection to PostgreSQL and runs AutoMigrate for all
// application models.
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("gorm open: %w", err)
	}

	// Run migrations.
	if err := db.AutoMigrate(&models.User{}, &models.Product{}); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	log.Println("database connected and migrated")
	return db, nil
}
