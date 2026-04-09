package middleware

import (
	"log"
	"runtime/debug"

	"github.com/gofiber/fiber/v2"
)

// Recovery returns a middleware that recovers from panics, logs the stack trace,
// and returns a 500 Internal Server Error JSON response.
func Recovery() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic recovered: %v\nStack: %s", r, debug.Stack())
				_ = c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
				})
			}
		}()
		return c.Next()
	}
}
