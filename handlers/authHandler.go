package handlers

import (
	"os"
	"strings"

	"github.com/JimcostDev/finances-api/middleware"
	"github.com/JimcostDev/finances-api/services"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	service services.AuthService
	users   services.UserService
}

func NewAuthHandler(s services.AuthService, us services.UserService) *AuthHandler {
	return &AuthHandler{service: s, users: us}
}

// cookieSecure: HTTPS directo, proxy (Koyeb, etc.) o COOKIE_SECURE=true.
func cookieSecure(c *fiber.Ctx) bool {
	if c.Protocol() == "https" {
		return true
	}
	if strings.EqualFold(c.Get("X-Forwarded-Proto"), "https") {
		return true
	}
	return os.Getenv("COOKIE_SECURE") == "true"
}

// cookieSameSite: en local (HTTP, mismo "site" localhost) Lax basta.
// En producción, front y API suelen ser dominios distintos: hace falta None + Secure o el navegador no manda la cookie en fetch.
func cookieSameSite(secure bool) string {
	if secure {
		return "None"
	}
	return "Lax"
}

// Register maneja la solicitud de registro
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req services.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Error al parsear JSON"})
	}

	user, err := h.service.RegisterUser(c.Context(), req)
	if err != nil {
		if err.Error() == "las contraseñas no coinciden" || err.Error() == "el email o el username ya existen" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(user)
}

// Login maneja la solicitud de inicio de sesión
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req services.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Error al parsear JSON"})
	}

	token, err := h.service.LoginUser(c.Context(), req)
	if err != nil {
		if err.Error() == "credenciales inválidas" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	secure := cookieSecure(c)
	c.Cookie(&fiber.Cookie{
		Name:     middleware.AuthCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   30 * 24 * 60 * 60,
		HTTPOnly: true,
		Secure:   secure,
		SameSite: cookieSameSite(secure),
	})

	return c.JSON(fiber.Map{"message": "Sesión iniciada"})
}

// Me devuelve el perfil del usuario autenticado (cookie o Bearer).
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Usuario no autenticado"})
	}
	user, err := h.users.GetUserProfile(c.Context(), userIDStr)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(user)
}

// Logout borra la cookie de sesión (no requiere JWT válido).
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	secure := cookieSecure(c)
	c.Cookie(&fiber.Cookie{
		Name:     middleware.AuthCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HTTPOnly: true,
		Secure:   secure,
		SameSite: cookieSameSite(secure),
	})
	return c.JSON(fiber.Map{"message": "Sesión cerrada"})
}
