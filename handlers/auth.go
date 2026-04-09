package handlers

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/mortogo321/go-fiber-api/config"
	"github.com/mortogo321/go-fiber-api/models"
	"github.com/mortogo321/go-fiber-api/utils"
)

var validate = validator.New()

// AuthHandler groups all authentication-related handlers.
type AuthHandler struct {
	DB  *gorm.DB
	Cfg *config.Config
}

// NewAuthHandler returns a ready-to-use AuthHandler.
func NewAuthHandler(db *gorm.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{DB: db, Cfg: cfg}
}

// RegisterInput represents the expected JSON body for registration.
type RegisterInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required"`
}

// LoginInput represents the expected JSON body for login.
type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// Register creates a new user account, hashes the password with bcrypt, and
// returns a signed JWT.
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var input RegisterInput
	if err := c.BodyParser(&input); err != nil {
		return utils.Error(c, http.StatusBadRequest, "invalid request body")
	}

	if err := validate.Struct(input); err != nil {
		return utils.Error(c, http.StatusBadRequest, err.Error())
	}

	// Check for existing user.
	var existing models.User
	if err := h.DB.Where("email = ?", input.Email).First(&existing).Error; err == nil {
		return utils.Error(c, http.StatusConflict, "email already registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return utils.Error(c, http.StatusInternalServerError, "failed to hash password")
	}

	user := models.User{
		Email:    input.Email,
		Password: string(hashedPassword),
		Name:     input.Name,
		Role:     "user",
	}

	if err := h.DB.Create(&user).Error; err != nil {
		return utils.Error(c, http.StatusInternalServerError, "failed to create user")
	}

	token, err := utils.GenerateToken(user.ID, user.Role, h.Cfg.JWTSecret, h.Cfg.JWTExpiry)
	if err != nil {
		return utils.Error(c, http.StatusInternalServerError, "failed to generate token")
	}

	return utils.Success(c, fiber.Map{
		"token": token,
		"user": fiber.Map{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
			"role":  user.Role,
		},
	})
}

// Login authenticates a user by email and password and returns a signed JWT.
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		return utils.Error(c, http.StatusBadRequest, "invalid request body")
	}

	if err := validate.Struct(input); err != nil {
		return utils.Error(c, http.StatusBadRequest, err.Error())
	}

	var user models.User
	if err := h.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return utils.Error(c, http.StatusUnauthorized, "invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return utils.Error(c, http.StatusUnauthorized, "invalid credentials")
	}

	token, err := utils.GenerateToken(user.ID, user.Role, h.Cfg.JWTSecret, h.Cfg.JWTExpiry)
	if err != nil {
		return utils.Error(c, http.StatusInternalServerError, "failed to generate token")
	}

	return utils.Success(c, fiber.Map{
		"token": token,
		"user": fiber.Map{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
			"role":  user.Role,
		},
	})
}

// GetProfile returns the currently authenticated user's profile based on the
// JWT claims stored in Fiber locals by the auth middleware.
func (h *AuthHandler) GetProfile(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uint)
	if !ok {
		return utils.Error(c, http.StatusUnauthorized, "unauthorized")
	}

	var user models.User
	if err := h.DB.First(&user, userID).Error; err != nil {
		return utils.Error(c, http.StatusNotFound, "user not found")
	}

	return utils.Success(c, fiber.Map{
		"id":         user.ID,
		"email":      user.Email,
		"name":       user.Name,
		"role":       user.Role,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	})
}
