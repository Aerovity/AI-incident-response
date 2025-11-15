package middleware

import (
	"incident-backend/internal/models"
	"incident-backend/internal/services"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware creates a middleware for JWT authentication
func AuthMiddleware(authService *services.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing authorization header",
				"code":  401,
			})
		}

		// Check if it's a Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format",
				"code":  401,
			})
		}

		tokenString := parts[1]

		// Validate token
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
				"code":  401,
			})
		}

		// Set claims in context
		c.Locals("userID", claims.UserID)
		c.Locals("organizationID", claims.OrganizationID)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}

// RequireRole creates a middleware that checks for specific roles
func RequireRole(roles ...models.UserRole) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole := c.Locals("role")
		if userRole == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
				"code":  401,
			})
		}

		role, ok := userRole.(models.UserRole)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
				"code":  500,
			})
		}

		// Check if user has required role
		for _, r := range roles {
			if role == r {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Insufficient permissions",
			"code":  403,
		})
	}
}
