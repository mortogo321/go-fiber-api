package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

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
	// ── Configuration ──────────────────────────────────────────────────
	cfg := config.LoadConfig()

	// ── Database ───────────────────────────────────────────────────────
	if err := database.InitDB(cfg); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	// ── Redis ──────────────────────────────────────────────────────────
	if err := database.InitRedis(cfg); err != nil {
		log.Fatalf("Redis initialization failed: %v", err)
	}

	// ── Services ───────────────────────────────────────────────────────
	cacheService := services.NewCacheService(database.RedisClient)

	// ── Handlers ───────────────────────────────────────────────────────
	authHandler := handlers.NewAuthHandler(cfg.JWTSecret)
	productHandler := handlers.NewProductHandler(cacheService)

	// ── Fiber App ──────────────────────────────────────────────────────
	app := fiber.New(fiber.Config{
		AppName:      "Go Fiber API",
		ErrorHandler: customErrorHandler,
	})

	// Global middleware.
	app.Use(recover.New())
	app.Use(cors.New())
	app.Use(middleware.RequestLogger())

	// ── Health Check ───────────────────────────────────────────────────
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// ── API Routes ─────────────────────────────────────────────────────
	api := app.Group("/api")

	// Public auth routes.
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

	// Protected routes — require JWT.
	protected := api.Group("", middleware.JWTMiddleware(cfg.JWTSecret))

	// User routes.
	users := protected.Group("/users")
	users.Get("/profile", authHandler.GetProfile)

	// Product routes.
	products := protected.Group("/products")
	products.Get("/", productHandler.GetProducts)
	products.Get("/:id", productHandler.GetProduct)
	products.Post("/", productHandler.CreateProduct)
	products.Put("/:id", productHandler.UpdateProduct)
	products.Delete("/:id", productHandler.DeleteProduct)

	// ── Graceful Shutdown ──────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	log.Printf("Server is running on port %s", cfg.Port)

	<-quit
	log.Println("Shutting down server...")

	if err := app.Shutdown(); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped gracefully")
}

// customErrorHandler returns a consistent JSON error for unhandled Fiber errors.
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
