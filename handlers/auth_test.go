package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/mortogo321/go-fiber-api/config"
	"github.com/mortogo321/go-fiber-api/models"
)

// setupTestDB creates a connection to the test database and auto-migrates.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	cfg := config.Load()
	dsn := "host=" + cfg.DBHost +
		" user=" + cfg.DBUser +
		" password=" + cfg.DBPassword +
		" dbname=" + cfg.DBName +
		" port=" + cfg.DBPort +
		" sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("skipping: cannot connect to test database: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Product{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	// Clean tables before each test.
	db.Exec("DELETE FROM products")
	db.Exec("DELETE FROM users")
	return db
}

func setupAuthApp(t *testing.T) (*fiber.App, *gorm.DB) {
	t.Helper()
	db := setupTestDB(t)
	cfg := &config.Config{JWTSecret: "test-secret"}
	h := NewAuthHandler(db, cfg)
	app := fiber.New()
	app.Post("/register", h.Register)
	app.Post("/login", h.Login)
	return app, db
}

func parseResponseBody(t *testing.T, resp *http.Response) map[string]interface{} {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	return result
}

func TestRegister_Success(t *testing.T) {
	app, _ := setupAuthApp(t)

	reqBody := RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusCreated {
		t.Errorf("expected status %d, got %d", fiber.StatusCreated, resp.StatusCode)
	}

	result := parseResponseBody(t, resp)
	if result["success"] != true {
		t.Errorf("expected success true, got %v", result["success"])
	}
}

func TestRegister_DuplicateEmail_Returns409(t *testing.T) {
	app, db := setupAuthApp(t)

	// Seed a user directly in the DB.
	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	db.Create(&models.User{
		Email:    "dup@example.com",
		Password: string(hashed),
		Name:     "Existing",
		Role:     "user",
	})

	reqBody := RegisterRequest{
		Email:    "dup@example.com",
		Password: "password123",
		Name:     "Another",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusConflict {
		t.Errorf("expected status %d, got %d", fiber.StatusConflict, resp.StatusCode)
	}
}

func TestLogin_Success_ReturnsJWT(t *testing.T) {
	app, db := setupAuthApp(t)

	// Seed a user.
	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	db.Create(&models.User{
		Email:    "login@example.com",
		Password: string(hashed),
		Name:     "Login User",
		Role:     "user",
	})

	reqBody := LoginRequest{Email: "login@example.com", Password: "password123"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}

	result := parseResponseBody(t, resp)
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected data object in response")
	}

	tokenStr, ok := data["token"].(string)
	if !ok || tokenStr == "" {
		t.Error("expected non-empty token in response")
	}

	// Verify the token is valid JWT.
	_, err = jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})
	if err != nil {
		t.Errorf("token is not valid JWT: %v", err)
	}
}

func TestLogin_WrongPassword_Returns401(t *testing.T) {
	app, db := setupAuthApp(t)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	db.Create(&models.User{
		Email:    "wrongpw@example.com",
		Password: string(hashed),
		Name:     "User",
		Role:     "user",
	})

	reqBody := LoginRequest{Email: "wrongpw@example.com", Password: "wrong-password"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", fiber.StatusUnauthorized, resp.StatusCode)
	}
}

func TestLogin_NonExistentUser_Returns401(t *testing.T) {
	app, _ := setupAuthApp(t)

	reqBody := LoginRequest{Email: "nouser@example.com", Password: "password123"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", fiber.StatusUnauthorized, resp.StatusCode)
	}
}
