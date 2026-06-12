package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

const userContextKey = "auth_user"

func RequireRoles(manager *TokenManager, allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := strings.TrimSpace(c.Get(fiber.HeaderAuthorization))
		if header == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "требуется авторизация"})
		}

		if !strings.HasPrefix(strings.ToLower(header), "bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "неверный формат токена"})
		}

		token := strings.TrimSpace(header[len("Bearer "):])
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "требуется авторизация"})
		}

		user, err := manager.ParseToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}

		c.Locals(userContextKey, user)

		if len(allowedRoles) == 0 {
			return c.Next()
		}

		for _, role := range allowedRoles {
			if user.Role == role {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "недостаточно прав для выполнения операции"})
	}
}

func CurrentUser(c *fiber.Ctx) (User, bool) {
	user, ok := c.Locals(userContextKey).(User)
	return user, ok
}
