package middleware

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

// RequestLogger returns a Fiber middleware that logs the HTTP method, path,
// response status code, and request latency for every request.
func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request.
		err := c.Next()

		latency := time.Since(start)
		log.Printf("%s %s %d %v",
			c.Method(),
			c.Path(),
			c.Response().StatusCode(),
			latency,
		)

		return err
	}
}
