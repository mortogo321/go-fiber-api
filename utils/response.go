package utils

import "github.com/gofiber/fiber/v2"

// Success sends a standardized success JSON response.
func Success(c *fiber.Ctx, data interface{}) error {
	return c.JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

// Error sends a standardized error JSON response with the given HTTP status.
func Error(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"success": false,
		"error":   message,
	})
}
