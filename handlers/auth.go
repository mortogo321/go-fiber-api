package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mortogo321/go-fiber-api/database"
	"github.com/mortogo321/go-fiber-api/models"
	"github.com/mortogo321/go-fiber-api/utils"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler holds dependencies for authentication endpoints.
type AuthHandler struct {
	JWTSecret string
}

// NewAuthHandler creates an AuthHandler with the provided JWT secret.
func NewAuthHandler(jwtSecret string) *AuthHandler {
	return &AuthHandler{JWTSecret: jwtSecret}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register creates a new user account and returns a JWT token.
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req registerRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		return utils.Error(c, fiber.StatusBadRequest, "email, password, and name are required")
	}

	// Check for existing user.
	var existing models.User
	if result := database.DB.Where("email = ?", req.Email).First(&existing); result.Error == nil {
		return utils.Error(c, fiber.StatusConflict, "email already registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "failed to hash password")
	}

	user := models.User{
		Email:    req.Email,
		Password: string(hashedPassword),
		Name:     req.Name,
		Role:     "user",
	}

	if result := database.DB.Create(&user); result.Error != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "failed to create user")
	}

	token, err := utils.GenerateToken(user.ID, user.Role, h.JWTSecret)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "failed to generate token")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"token": token,
			"user":  user,
		},
	})
}

// Login authenticates a user by email/password and returns a JWT token.
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	if req.Email == "" || req.Password == "" {
		return utils.Error(c, fiber.StatusBadRequest, "email and password are required")
	}

	var user models.User
	if result := database.DB.Where("email = ?", req.Email).First(&user); result.Error != nil {
		return utils.Error(c, fiber.StatusUnauthorized, "invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return utils.Error(c, fiber.StatusUnauthorized, "invalid credentials")
	}

	token, err := utils.GenerateToken(user.ID, user.Role, h.JWTSecret)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "failed to generate token")
	}

	return utils.Success(c, fiber.Map{
		"token": token,
		"user":  user,
	})
}

// GetProfile returns the authenticated user's profile.
func (h *AuthHandler) GetProfile(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uint)
	if !ok {
		return utils.Error(c, fiber.StatusUnauthorized, "invalid user context")
	}

	var user models.User
	if result := database.DB.First(&user, userID); result.Error != nil {
		return utils.Error(c, fiber.StatusNotFound, "user not found")
	}

	return utils.Success(c, user)
}
