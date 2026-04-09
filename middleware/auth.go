package middleware

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/mortogo321/go-fiber-api/utils"
)

// JWTMiddleware returns a Fiber handler that extracts and validates a Bearer
// token from the Authorization header. On success it stores userID and role in
// c.Locals so downstream handlers can access them.
func JWTMiddleware(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return utils.Error(c, http.StatusUnauthorized, "missing authorization header")
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return utils.Error(c, http.StatusUnauthorized, "invalid authorization format")
		}

		claims, err := utils.ValidateToken(parts[1], secret)
		if err != nil {
			return utils.Error(c, http.StatusUnauthorized, "invalid or expired token")
		}

		// Store claims in Fiber locals for downstream handlers.
		c.Locals("userID", claims.UserID)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}
