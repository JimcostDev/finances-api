package handlers

import (
	"github.com/JimcostDev/finances-api/services"
	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	service services.UserService
}

func NewUserHandler(s services.UserService) *UserHandler {
	return &UserHandler{service: s}
}

func (h *UserHandler) GetUserProfile(c *fiber.Ctx) error {
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Usuario no autenticado"})
	}

	user, err := h.service.GetUserProfile(c.Context(), userIDStr)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(user)
}

func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Usuario no autenticado"})
	}

	var req services.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Error al parsear JSON"})
	}

	err := h.service.UpdateUser(c.Context(), userIDStr, req)
	if err != nil {
		// Podríamos afinar los status codes según el error (conflict vs internal),
		// pero por simplicidad generalizamos o chequeamos mensajes string.
		if err.Error() == "el email ya está en uso" || err.Error() == "el nombre de usuario ya está en uso" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		}
		if err.Error() == "ID inválido" || err.Error() == "las contraseñas no coinciden" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Usuario actualizado correctamente"})
}

func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Usuario no autenticado"})
	}

	err := h.service.DeleteUser(c.Context(), userIDStr)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Usuario y reportes asociados eliminados exitosamente"})
}
