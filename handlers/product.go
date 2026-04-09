package handlers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/mor-tesla/go-fiber-api/models"
	"github.com/mor-tesla/go-fiber-api/services"
	"github.com/mor-tesla/go-fiber-api/utils"
)

type ProductHandler struct {
	db    *gorm.DB
	cache *services.CacheService
}

func NewProductHandler(db *gorm.DB, cache *services.CacheService) *ProductHandler {
	return &ProductHandler{db: db, cache: cache}
}

type CreateProductRequest struct {
	Name        string  `json:"name" validate:"required,min=2"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	SKU         string  `json:"sku" validate:"required,min=3"`
}

type UpdateProductRequest struct {
	Name        *string  `json:"name" validate:"omitempty,min=2"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price" validate:"omitempty,gt=0"`
	SKU         *string  `json:"sku" validate:"omitempty,min=3"`
}

const (
	productsCacheKey = "products:all"
	productCacheKey  = "products:%d"
	cacheTTL         = 5 * time.Minute
)

// GetProducts returns all products with Redis cache-aside pattern.
func (h *ProductHandler) GetProducts(c *fiber.Ctx) error {
	// Try cache first
	cached, err := h.cache.Get(c.Context(), productsCacheKey)
	if err == nil && cached != "" {
		var products []models.Product
		if err := json.Unmarshal([]byte(cached), &products); err == nil {
			return utils.SuccessResponse(c, fiber.StatusOK, "products retrieved (cached)", products)
		}
	}

	// Fallback to database
	var products []models.Product
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int64
	h.db.Model(&models.Product{}).Count(&total)

	if result := h.db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&products); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to fetch products")
	}

	// Populate cache
	if data, err := json.Marshal(products); err == nil {
		_ = h.cache.Set(c.Context(), productsCacheKey, string(data), cacheTTL)
	}

	return utils.PaginateResponse(c, fiber.StatusOK, "products retrieved", products, page, limit, total)
}

// GetProduct returns a single product by ID with caching.
func (h *ProductHandler) GetProduct(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid product ID")
	}

	cacheKey := fmt.Sprintf(productCacheKey, id)

	// Try cache first
	cached, err := h.cache.Get(c.Context(), cacheKey)
	if err == nil && cached != "" {
		var product models.Product
		if err := json.Unmarshal([]byte(cached), &product); err == nil {
			return utils.SuccessResponse(c, fiber.StatusOK, "product retrieved (cached)", product)
		}
	}

	var product models.Product
	if result := h.db.First(&product, id); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "product not found")
	}

	// Populate cache
	if data, err := json.Marshal(product); err == nil {
		_ = h.cache.Set(c.Context(), cacheKey, string(data), cacheTTL)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "product retrieved", product)
}

// CreateProduct creates a new product for the authenticated user.
func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
	var req CreateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid request body")
	}

	if errs := utils.ValidateStruct(req); errs != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"success": false,
			"errors":  errs,
		})
	}

	userID, ok := c.Locals("user_id").(float64)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "unauthorized")
	}

	product := models.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		SKU:         req.SKU,
		UserID:      uint(userID),
	}

	if result := h.db.Create(&product); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to create product")
	}

	// Invalidate products list cache
	_ = h.cache.InvalidatePattern(c.Context(), "products:*")

	return utils.SuccessResponse(c, fiber.StatusCreated, "product created", product)
}

// UpdateProduct updates an existing product owned by the authenticated user.
func (h *ProductHandler) UpdateProduct(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid product ID")
	}

	var req UpdateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid request body")
	}

	if errs := utils.ValidateStruct(req); errs != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"success": false,
			"errors":  errs,
		})
	}

	userID, ok := c.Locals("user_id").(float64)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "unauthorized")
	}

	var product models.Product
	if result := h.db.First(&product, id); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "product not found")
	}

	if product.UserID != uint(userID) {
		return utils.ErrorResponse(c, fiber.StatusForbidden, "not authorized to update this product")
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Price != nil {
		updates["price"] = *req.Price
	}
	if req.SKU != nil {
		updates["sku"] = *req.SKU
	}

	if result := h.db.Model(&product).Updates(updates); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to update product")
	}

	// Invalidate relevant caches
	_ = h.cache.InvalidatePattern(c.Context(), "products:*")

	return utils.SuccessResponse(c, fiber.StatusOK, "product updated", product)
}

// DeleteProduct removes a product owned by the authenticated user.
func (h *ProductHandler) DeleteProduct(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid product ID")
	}

	userID, ok := c.Locals("user_id").(float64)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "unauthorized")
	}

	var product models.Product
	if result := h.db.First(&product, id); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "product not found")
	}

	if product.UserID != uint(userID) {
		return utils.ErrorResponse(c, fiber.StatusForbidden, "not authorized to delete this product")
	}

	if result := h.db.Delete(&product); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to delete product")
	}

	// Invalidate relevant caches
	_ = h.cache.InvalidatePattern(c.Context(), "products:*")

	return utils.SuccessResponse(c, fiber.StatusOK, "product deleted", nil)
}
