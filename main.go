package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/mor-tesla/go-fiber-api/config"
	"github.com/mor-tesla/go-fiber-api/database"
	"github.com/mor-tesla/go-fiber-api/handlers"
	"github.com/mor-tesla/go-fiber-api/middleware"
	"github.com/mor-tesla/go-fiber-api/services"
)

func main() {
	cfg := config.Load()

	db := database.ConnectPostgres(cfg)
	rdb := database.ConnectRedis(cfg)
	cache := services.NewCacheService(rdb)

	app := fiber.New(fiber.Config{
		AppName:      "Go Fiber API",
		ErrorHandler: customErrorHandler,
	})

	app.Use(recover.New())
	app.Use(cors.New())
	app.Use(middleware.LoggerMiddleware())

	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	authHandler := handlers.NewAuthHandler(db, cfg)
	productHandler := handlers.NewProductHandler(db, cache)

	auth := app.Group("/api/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Get("/profile", middleware.JWTMiddleware(cfg.JWTSecret), authHandler.GetProfile)

	products := app.Group("/api/products", middleware.JWTMiddleware(cfg.JWTSecret))
	products.Get("/", productHandler.GetProducts)
	products.Get("/:id", productHandler.GetProduct)
	products.Post("/", productHandler.CreateProduct)
	products.Put("/:id", productHandler.UpdateProduct)
	products.Delete("/:id", productHandler.DeleteProduct)

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

	if err := app.Shutdown(); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}

	if err := rdb.Close(); err != nil {
		log.Printf("redis close error: %v", err)
	}

	sqlDB, _ := db.DB()
	if sqlDB != nil {
		if err := sqlDB.Close(); err != nil {
			log.Printf("postgres close error: %v", err)
		}
	}

	log.Println("server stopped")
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"success": false,
		"error":   err.Error(),
	})
}
