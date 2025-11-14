package handlers

import (
	"incident-backend/internal/services"
	"incident-backend/internal/storage"

	"github.com/gofiber/fiber/v2"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *services.AuthService
	userStore   *storage.UserStore
	orgStore    *storage.OrgStore
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *services.AuthService, userStore *storage.UserStore, orgStore *storage.OrgStore) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userStore:   userStore,
		orgStore:    orgStore,
	}
}

// Register handles user registration
// @route POST /api/v1/auth/register
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req services.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"code":  400,
		})
	}

	authResp, err := h.authService.Register(&req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
			"code":  400,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(authResp)
}

// Login handles user login
// @route POST /api/v1/auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req services.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"code":  400,
		})
	}

	authResp, err := h.authService.Login(&req)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
			"code":  401,
		})
	}

	return c.JSON(authResp)
}

// GetMe returns the current user information
// @route GET /api/v1/auth/me
func (h *AuthHandler) GetMe(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	orgID := c.Locals("organizationID").(string)

	user, err := h.userStore.GetByID(orgID, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
			"code":  404,
		})
	}

	org, err := h.orgStore.GetByID(orgID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Organization not found",
			"code":  404,
		})
	}

	return c.JSON(fiber.Map{
		"user":         user.ToResponse(),
		"organization": org,
	})
}

// RefreshToken refreshes the JWT token
// @route POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	orgID := c.Locals("organizationID").(string)

	user, err := h.userStore.GetByID(orgID, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
			"code":  404,
		})
	}

	token, err := h.authService.GenerateToken(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
			"code":  500,
		})
	}

	return c.JSON(fiber.Map{
		"token": token,
	})
}
