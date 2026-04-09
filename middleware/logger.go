package middleware

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

// LoggerMiddleware logs each request with method, path, status code, and duration.
func LoggerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start)
		status := c.Response().StatusCode()

		log.Printf("%-7s %s -> %d (%s)",
			c.Method(),
			c.Path(),
			status,
			duration.Round(time.Microsecond),
		)

		return err
	}
}
