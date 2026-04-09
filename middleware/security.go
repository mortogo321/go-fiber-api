package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
)

// SecurityHeaders returns a middleware that sets common security HTTP headers
// (X-Content-Type-Options, X-Frame-Options, etc.) using Fiber's helmet middleware.
func SecurityHeaders() fiber.Handler {
	return helmet.New()
}
