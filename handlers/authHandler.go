package handlers

import (
	"github.com/JimcostDev/finances-api/services"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	service services.AuthService
}

func NewAuthHandler(s services.AuthService) *AuthHandler {
	return &AuthHandler{service: s}
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

	return c.JSON(fiber.Map{"token": token})
}
