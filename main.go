package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/mortogo321/go-fiber-api/config"
	"github.com/mortogo321/go-fiber-api/database"
	"github.com/mortogo321/go-fiber-api/handlers"
	"github.com/mortogo321/go-fiber-api/middleware"
	"github.com/mortogo321/go-fiber-api/services"
)

func main() {
	// Load configuration from environment variables.
	cfg := config.LoadConfig()

	// Initialize PostgreSQL via GORM.
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Initialize Redis client.
	rdb, err := database.InitRedis(cfg)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}

	// Build shared services.
	cacheService := services.NewCacheService(rdb)

	// Create Fiber app.
	app := fiber.New(fiber.Config{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	})

	// Global middleware.
	app.Use(recover.New())
	app.Use(cors.New())
	app.Use(middleware.RequestLogger())

	// Health check.
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// --- Auth routes (public) ---
	authHandler := handlers.NewAuthHandler(db, cfg)
	auth := app.Group("/api/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Get("/profile", middleware.JWTMiddleware(cfg.JWTSecret), authHandler.GetProfile)

	// --- Product routes (protected) ---
	productHandler := handlers.NewProductHandler(db, cacheService)
	products := app.Group("/api/products", middleware.JWTMiddleware(cfg.JWTSecret))
	products.Get("/", productHandler.GetProducts)
	products.Get("/:id", productHandler.GetProduct)
	products.Post("/", productHandler.CreateProduct)
	products.Put("/:id", productHandler.UpdateProduct)
	products.Delete("/:id", productHandler.DeleteProduct)

	// Graceful shutdown: listen for SIGINT / SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	log.Printf("server started on port %s", cfg.Port)

	<-quit
	log.Println("shutting down server...")

	if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
		log.Fatalf("server forced shutdown: %v", err)
	}

	// Close Redis connection.
	if err := rdb.Close(); err != nil {
		log.Printf("error closing redis: %v", err)
	}

	// Close database connection.
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		if err := sqlDB.Close(); err != nil {
			log.Printf("error closing database: %v", err)
		}
	}

	_ = context.Background() // satisfy import for future use
	log.Println("server exited gracefully")
}
