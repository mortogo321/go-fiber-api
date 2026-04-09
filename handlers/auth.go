package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/mor-tesla/go-fiber-api/config"
	"github.com/mor-tesla/go-fiber-api/models"
	"github.com/mor-tesla/go-fiber-api/utils"
)

type AuthHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewAuthHandler(db *gorm.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{db: db, cfg: cfg}
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required,min=2"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string              `json:"token"`
	User  models.UserResponse `json:"user"`
}

// Register creates a new user account with a hashed password.
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid request body")
	}

	if errs := utils.ValidateStruct(req); errs != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"success": false,
			"errors":  errs,
		})
	}

	var existing models.User
	if result := h.db.Where("email = ?", req.Email).First(&existing); result.Error == nil {
		return utils.ErrorResponse(c, fiber.StatusConflict, "email already registered")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to hash password")
	}

	user := models.User{
		Email:    req.Email,
		Password: string(hashed),
		Name:     req.Name,
		Role:     "user",
	}

	if result := h.db.Create(&user); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to create user")
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "user registered successfully", user.ToResponse())
}

// Login authenticates a user and returns a JWT token.
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid request body")
	}

	if errs := utils.ValidateStruct(req); errs != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"success": false,
			"errors":  errs,
		})
	}

	var user models.User
	if result := h.db.Where("email = ?", req.Email).First(&user); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "invalid credentials")
	}

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to generate token")
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "login successful", LoginResponse{
		Token: tokenString,
		User:  user.ToResponse(),
	})
}

// GetProfile returns the current authenticated user's profile.
func (h *AuthHandler) GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "unauthorized")
	}

	var user models.User
	if result := h.db.First(&user, userID); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "user not found")
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "profile retrieved", user.ToResponse())
}
