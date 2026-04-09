package middleware

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

// RequestLogger returns a Fiber handler that logs each request's method, path,
// status code, and latency.
func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		latency := time.Since(start)
		log.Printf("[%s] %s %d %v",
			c.Method(),
			c.Path(),
			c.Response().StatusCode(),
			latency,
		)

		return err
	}
}
