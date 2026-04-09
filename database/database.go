package database

import (
	"fmt"
	"log"

	"github.com/mor-tesla/go-fiber-api/config"
	"github.com/mor-tesla/go-fiber-api/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ConnectPostgres establishes a GORM connection to PostgreSQL and runs AutoMigrate.
func ConnectPostgres(cfg *config.Config) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	if err := db.AutoMigrate(&models.User{}, &models.Product{}); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	log.Println("connected to PostgreSQL")
	return db
}
