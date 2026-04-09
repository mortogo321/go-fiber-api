package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/mortogo321/go-fiber-api/database"
	"github.com/mortogo321/go-fiber-api/models"
	"github.com/mortogo321/go-fiber-api/services"
	"github.com/mortogo321/go-fiber-api/utils"
)

const (
	productsListKey = "products:list"
	productKeyFmt   = "products:%s"
	cacheTTL        = 5 * time.Minute
)

// ProductHandler holds dependencies for product CRUD endpoints.
type ProductHandler struct {
	Cache *services.CacheService
}

// NewProductHandler creates a ProductHandler with the given cache service.
func NewProductHandler(cache *services.CacheService) *ProductHandler {
	return &ProductHandler{Cache: cache}
}

type createProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	SKU         string  `json:"sku"`
}

type updateProductRequest struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price"`
}

// GetProducts returns all products, using Redis cache-aside pattern.
func (h *ProductHandler) GetProducts(c *fiber.Ctx) error {
	ctx := context.Background()

	// 1. Check cache.
	cached, err := h.Cache.Get(ctx, productsListKey)
	if err == nil && cached != "" {
		var products []models.Product
		if json.Unmarshal([]byte(cached), &products) == nil {
			return utils.Success(c, products)
		}
	}

	// 2. Cache miss — query DB.
	var products []models.Product
	if result := database.DB.Order("created_at DESC").Find(&products); result.Error != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "failed to fetch products")
	}

	// 3. Store in cache.
	if data, err := json.Marshal(products); err == nil {
		_ = h.Cache.Set(ctx, productsListKey, string(data), cacheTTL)
	}

	return utils.Success(c, products)
}

// GetProduct returns a single product by ID, using Redis cache-aside pattern.
func (h *ProductHandler) GetProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	ctx := context.Background()
	cacheKey := fmt.Sprintf(productKeyFmt, id)

	// 1. Check cache.
	cached, err := h.Cache.Get(ctx, cacheKey)
	if err == nil && cached != "" {
		var product models.Product
		if json.Unmarshal([]byte(cached), &product) == nil {
			return utils.Success(c, product)
		}
	}

	// 2. Cache miss — query DB.
	var product models.Product
	if result := database.DB.First(&product, id); result.Error != nil {
		return utils.Error(c, fiber.StatusNotFound, "product not found")
	}

	// 3. Store in cache.
	if data, err := json.Marshal(product); err == nil {
		_ = h.Cache.Set(ctx, cacheKey, string(data), cacheTTL)
	}

	return utils.Success(c, product)
}

// CreateProduct creates a new product and invalidates the list cache.
func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
	var req createProductRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	if req.Name == "" || req.SKU == "" || req.Price <= 0 {
		return utils.Error(c, fiber.StatusBadRequest, "name, sku, and a positive price are required")
	}

	userID, ok := c.Locals("userID").(uint)
	if !ok {
		return utils.Error(c, fiber.StatusUnauthorized, "invalid user context")
	}

	product := models.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		SKU:         req.SKU,
		UserID:      userID,
	}

	if result := database.DB.Create(&product); result.Error != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "failed to create product")
	}

	// Invalidate list cache so the next read picks up the new product.
	_ = h.Cache.Delete(context.Background(), productsListKey)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    product,
	})
}

// UpdateProduct updates an existing product and invalidates related caches.
func (h *ProductHandler) UpdateProduct(c *fiber.Ctx) error {
	id := c.Params("id")

	var product models.Product
	if result := database.DB.First(&product, id); result.Error != nil {
		return utils.Error(c, fiber.StatusNotFound, "product not found")
	}

	var req updateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	if req.Name != nil {
		product.Name = *req.Name
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.Price != nil {
		if *req.Price <= 0 {
			return utils.Error(c, fiber.StatusBadRequest, "price must be positive")
		}
		product.Price = *req.Price
	}

	if result := database.DB.Save(&product); result.Error != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "failed to update product")
	}

	// Invalidate both the specific product cache and the list cache.
	ctx := context.Background()
	_ = h.Cache.Delete(ctx, fmt.Sprintf(productKeyFmt, id))
	_ = h.Cache.Delete(ctx, productsListKey)

	return utils.Success(c, product)
}

// DeleteProduct removes a product and invalidates related caches.
func (h *ProductHandler) DeleteProduct(c *fiber.Ctx) error {
	id := c.Params("id")

	var product models.Product
	if result := database.DB.First(&product, id); result.Error != nil {
		return utils.Error(c, fiber.StatusNotFound, "product not found")
	}

	if result := database.DB.Delete(&product); result.Error != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "failed to delete product")
	}

	// Invalidate caches.
	ctx := context.Background()
	_ = h.Cache.Delete(ctx, fmt.Sprintf(productKeyFmt, id))
	_ = h.Cache.Delete(ctx, productsListKey)

	return utils.Success(c, fiber.Map{"message": "product deleted"})
}
