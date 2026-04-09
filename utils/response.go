package utils

import "github.com/gofiber/fiber/v2"

// SuccessResponse sends a standardized success JSON response.
func SuccessResponse(c *fiber.Ctx, status int, message string, data interface{}) error {
	return c.Status(status).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
	})
}

// ErrorResponse sends a standardized error JSON response.
func ErrorResponse(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"success": false,
		"error":   message,
	})
}

// PaginateResponse sends a standardized paginated JSON response.
func PaginateResponse(c *fiber.Ctx, status int, message string, data interface{}, page, limit int, total int64) error {
	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return c.Status(status).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
		"meta": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}
