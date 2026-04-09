package utils

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// Success sends a 200 JSON response with the standard envelope.
func Success(c *fiber.Ctx, data interface{}) error {
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

// Error sends an error JSON response with the given HTTP status code.
func Error(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"success": false,
		"error":   message,
	})
}

// Paginated sends a 200 JSON response that includes pagination metadata.
func Paginated(c *fiber.Ctx, data interface{}, total int64, page int, limit int) error {
	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    data,
		"meta": fiber.Map{
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": totalPages,
		},
	})
}
