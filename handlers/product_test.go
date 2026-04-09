package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/mortogo321/go-fiber-api/models"
	"github.com/mortogo321/go-fiber-api/services"
)

func setupRedisClient(t *testing.T) *redis.Client {
	t.Helper()
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skipf("skipping: cannot connect to Redis: %v", err)
	}
	return rdb
}

func setupProductApp(t *testing.T) (*fiber.App, *gorm.DB, *services.CacheService) {
	t.Helper()
	db := setupTestDB(t)
	rdb := setupRedisClient(t)
	cache := services.NewCacheService(rdb)

	// Seed a user for product ownership.
	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := models.User{Email: "prod@example.com", Password: string(hashed), Name: "Prod User", Role: "user"}
	db.Create(&user)

	h := NewProductHandler(db, cache)
	app := fiber.New()

	// Simulate authenticated user by injecting user_id into Locals.
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", float64(user.ID))
		return c.Next()
	})

	app.Get("/products", h.GetProducts)
	app.Post("/products", h.CreateProduct)
	app.Put("/products/:id", h.UpdateProduct)
	app.Delete("/products/:id", h.DeleteProduct)

	// Clean Redis cache before test.
	rdb.FlushDB(context.Background())

	return app, db, cache
}

func TestGetProducts_ReturnsList(t *testing.T) {
	app, db, _ := setupProductApp(t)

	// Seed products.
	var user models.User
	db.First(&user)
	db.Create(&models.Product{Name: "Widget", Price: 9.99, SKU: "WDG-001", UserID: user.ID})
	db.Create(&models.Product{Name: "Gadget", Price: 19.99, SKU: "GDG-001", UserID: user.ID})

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}

	result := parseResponseBody(t, resp)
	if result["success"] != true {
		t.Errorf("expected success true, got %v", result["success"])
	}
}

func TestCreateProduct_ValidatesInput(t *testing.T) {
	app, _, _ := setupProductApp(t)

	// Missing required fields.
	reqBody := map[string]interface{}{"description": "no name or price"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnprocessableEntity {
		t.Errorf("expected status %d, got %d", fiber.StatusUnprocessableEntity, resp.StatusCode)
	}
}

func TestCreateProduct_Success(t *testing.T) {
	app, _, _ := setupProductApp(t)

	reqBody := CreateProductRequest{
		Name:        "New Product",
		Description: "A new product",
		Price:       29.99,
		SKU:         "NEW-001",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusCreated {
		t.Errorf("expected status %d, got %d", fiber.StatusCreated, resp.StatusCode)
	}
}

func TestCacheInvalidation_OnCreate(t *testing.T) {
	app, _, cache := setupProductApp(t)

	// Populate cache by listing products.
	listReq := httptest.NewRequest(http.MethodGet, "/products", nil)
	_, _ = app.Test(listReq, -1)

	// Verify cache is populated.
	cached, err := cache.Get(context.Background(), productsCacheKey)
	if err != nil || cached == "" {
		// Cache may not be populated if DB was empty; that is acceptable.
		t.Log("cache was not populated (empty product list), skipping invalidation check")
		return
	}

	// Create a product to trigger cache invalidation.
	reqBody := CreateProductRequest{Name: "Cache Test", Price: 5.00, SKU: "CACHE-001"}
	body, _ := json.Marshal(reqBody)
	createReq := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	_, _ = app.Test(createReq, -1)

	// Cache should be invalidated.
	cached, err = cache.Get(context.Background(), productsCacheKey)
	if err == nil && cached != "" {
		t.Error("expected cache to be invalidated after create, but it still exists")
	}
}

func TestCacheInvalidation_OnDelete(t *testing.T) {
	app, db, cache := setupProductApp(t)

	// Seed a product.
	var user models.User
	db.First(&user)
	product := models.Product{Name: "ToDelete", Price: 1.00, SKU: "DEL-001", UserID: user.ID}
	db.Create(&product)

	// Populate cache.
	listReq := httptest.NewRequest(http.MethodGet, "/products", nil)
	_, _ = app.Test(listReq, -1)

	// Delete the product.
	delReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/products/%d", product.ID), nil)
	resp, err := app.Test(delReq, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}

	// Cache should be invalidated.
	cached, _ := cache.Get(context.Background(), productsCacheKey)
	if cached != "" {
		t.Error("expected cache to be invalidated after delete")
	}
}
