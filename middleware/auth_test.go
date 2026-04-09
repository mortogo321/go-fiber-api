package middleware

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

const testJWTSecret = "test-secret-key"

func generateTestToken(t *testing.T, secret string, userID float64, role string, expiry time.Duration) string {
	t.Helper()
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(expiry).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return tokenString
}

func setupAuthTestApp() *fiber.App {
	app := fiber.New()
	app.Use(JWTMiddleware(testJWTSecret))
	app.Get("/protected", func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		role := c.Locals("role")
		return c.JSON(fiber.Map{
			"user_id": userID,
			"role":    role,
		})
	})
	return app
}

func TestJWTMiddleware_MissingAuthHeader_Returns401(t *testing.T) {
	app := setupAuthTestApp()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", fiber.StatusUnauthorized, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if result["success"] != false {
		t.Errorf("expected success false, got %v", result["success"])
	}
}

func TestJWTMiddleware_InvalidToken_Returns401(t *testing.T) {
	app := setupAuthTestApp()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-string")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", fiber.StatusUnauthorized, resp.StatusCode)
	}
}

func TestJWTMiddleware_ExpiredToken_Returns401(t *testing.T) {
	app := setupAuthTestApp()

	// Generate a token that expired 1 hour ago.
	expiredToken := generateTestToken(t, testJWTSecret, 1, "user", -1*time.Hour)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", fiber.StatusUnauthorized, resp.StatusCode)
	}
}

func TestJWTMiddleware_ValidToken_SetsUserIDInLocals(t *testing.T) {
	app := setupAuthTestApp()

	validToken := generateTestToken(t, testJWTSecret, 42, "admin", 1*time.Hour)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	userID, ok := result["user_id"].(float64)
	if !ok || userID != 42 {
		t.Errorf("expected user_id 42, got %v", result["user_id"])
	}

	role, ok := result["role"].(string)
	if !ok || role != "admin" {
		t.Errorf("expected role 'admin', got %v", result["role"])
	}
}

func TestJWTMiddleware_WrongSecret_Returns401(t *testing.T) {
	app := setupAuthTestApp()

	// Sign with a different secret.
	tokenWithWrongSecret := generateTestToken(t, "different-secret", 1, "user", 1*time.Hour)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenWithWrongSecret)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", fiber.StatusUnauthorized, resp.StatusCode)
	}
}

func TestJWTMiddleware_MalformedAuthHeader_Returns401(t *testing.T) {
	app := setupAuthTestApp()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "NotBearer sometoken")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	// The middleware checks for "Bearer" prefix; "NotBearer" should still
	// attempt to parse but fail since it is not a valid JWT.
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", fiber.StatusUnauthorized, resp.StatusCode)
	}
}
