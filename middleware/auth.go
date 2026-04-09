package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/mortogo321/go-fiber-api/utils"
)

// JWTMiddleware returns a Fiber handler that validates Bearer tokens in the
// Authorization header and sets "userID" and "role" in c.Locals for downstream handlers.
func JWTMiddleware(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return utils.Error(c, fiber.StatusUnauthorized, "missing authorization header")
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return utils.Error(c, fiber.StatusUnauthorized, "invalid authorization format")
		}

		claims, err := utils.ValidateToken(parts[1], jwtSecret)
		if err != nil {
			return utils.Error(c, fiber.StatusUnauthorized, "invalid or expired token")
		}

		c.Locals("userID", claims.UserID)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}
