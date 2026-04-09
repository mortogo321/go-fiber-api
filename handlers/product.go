package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/mortogo321/go-fiber-api/models"
	"github.com/mortogo321/go-fiber-api/services"
	"github.com/mortogo321/go-fiber-api/utils"
)

const (
	productsCacheKey = "products:list"
	productCacheKey  = "products:%d"
	productCacheTTL  = 5 * time.Minute
)

// ProductHandler groups product CRUD handlers.
type ProductHandler struct {
	DB    *gorm.DB
	Cache *services.CacheService
}

// NewProductHandler returns a ready-to-use ProductHandler.
func NewProductHandler(db *gorm.DB, cache *services.CacheService) *ProductHandler {
	return &ProductHandler{DB: db, Cache: cache}
}

// CreateProductInput represents the expected JSON body for product creation.
type CreateProductInput struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	SKU         string  `json:"sku"`
}

// UpdateProductInput represents the expected JSON body for product updates.
type UpdateProductInput struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price" validate:"omitempty,gt=0"`
	SKU         *string  `json:"sku"`
}

// GetProducts returns a paginated list of products. It checks Redis first
// (cache-aside); on a miss it queries PostgreSQL and populates the cache.
func (h *ProductHandler) GetProducts(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	cacheKey := fmt.Sprintf("%s:page:%d:limit:%d", productsCacheKey, page, limit)

	// Cache-aside: try Redis first.
	cached, err := h.Cache.Get(cacheKey)
	if err == nil && cached != "" {
		var result fiber.Map
		if json.Unmarshal([]byte(cached), &result) == nil {
			return utils.Success(c, result)
		}
	}

	// Cache miss -- query database.
	var products []models.Product
	var total int64

	h.DB.Model(&models.Product{}).Count(&total)
	h.DB.Preload("User").Offset(offset).Limit(limit).Order("id desc").Find(&products)

	response := fiber.Map{
		"products": products,
		"total":    total,
		"page":     page,
		"limit":    limit,
	}

	// Populate cache.
	if data, err := json.Marshal(response); err == nil {
		_ = h.Cache.Set(cacheKey, string(data), productCacheTTL)
	}

	return utils.Paginated(c, products, total, page, limit)
}

// GetProduct returns a single product by ID with cache-aside.
func (h *ProductHandler) GetProduct(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.Error(c, http.StatusBadRequest, "invalid product id")
	}

	cacheKey := fmt.Sprintf(productCacheKey, id)

	// Cache-aside: try Redis first.
	cached, cacheErr := h.Cache.Get(cacheKey)
	if cacheErr == nil && cached != "" {
		var product models.Product
		if json.Unmarshal([]byte(cached), &product) == nil {
			return utils.Success(c, product)
		}
	}

	var product models.Product
	if err := h.DB.Preload("User").First(&product, id).Error; err != nil {
		return utils.Error(c, http.StatusNotFound, "product not found")
	}

	// Populate cache.
	if data, err := json.Marshal(product); err == nil {
		_ = h.Cache.Set(cacheKey, string(data), productCacheTTL)
	}

	return utils.Success(c, product)
}

// CreateProduct creates a new product and invalidates the list cache.
func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
	var input CreateProductInput
	if err := c.BodyParser(&input); err != nil {
		return utils.Error(c, http.StatusBadRequest, "invalid request body")
	}

	if err := validate.Struct(input); err != nil {
		return utils.Error(c, http.StatusBadRequest, err.Error())
	}

	userID, _ := c.Locals("userID").(uint)

	product := models.Product{
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		SKU:         input.SKU,
		UserID:      userID,
	}

	if err := h.DB.Create(&product).Error; err != nil {
		return utils.Error(c, http.StatusInternalServerError, "failed to create product")
	}

	// Invalidate list cache.
	_ = h.Cache.DeleteByPattern("products:list*")

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    product,
	})
}

// UpdateProduct updates an existing product and invalidates relevant caches.
func (h *ProductHandler) UpdateProduct(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.Error(c, http.StatusBadRequest, "invalid product id")
	}

	var product models.Product
	if err := h.DB.First(&product, id).Error; err != nil {
		return utils.Error(c, http.StatusNotFound, "product not found")
	}

	var input UpdateProductInput
	if err := c.BodyParser(&input); err != nil {
		return utils.Error(c, http.StatusBadRequest, "invalid request body")
	}

	if err := validate.Struct(input); err != nil {
		return utils.Error(c, http.StatusBadRequest, err.Error())
	}

	if input.Name != nil {
		product.Name = *input.Name
	}
	if input.Description != nil {
		product.Description = *input.Description
	}
	if input.Price != nil {
		product.Price = *input.Price
	}
	if input.SKU != nil {
		product.SKU = *input.SKU
	}

	if err := h.DB.Save(&product).Error; err != nil {
		return utils.Error(c, http.StatusInternalServerError, "failed to update product")
	}

	// Invalidate both item and list caches.
	_ = h.Cache.Delete(fmt.Sprintf(productCacheKey, id))
	_ = h.Cache.DeleteByPattern("products:list*")

	return utils.Success(c, product)
}

// DeleteProduct removes a product and invalidates relevant caches.
func (h *ProductHandler) DeleteProduct(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.Error(c, http.StatusBadRequest, "invalid product id")
	}

	var product models.Product
	if err := h.DB.First(&product, id).Error; err != nil {
		return utils.Error(c, http.StatusNotFound, "product not found")
	}

	if err := h.DB.Delete(&product).Error; err != nil {
		return utils.Error(c, http.StatusInternalServerError, "failed to delete product")
	}

	// Invalidate both item and list caches.
	_ = h.Cache.Delete(fmt.Sprintf(productCacheKey, id))
	_ = h.Cache.DeleteByPattern("products:list*")

	return utils.Success(c, fiber.Map{"message": "product deleted"})
}
