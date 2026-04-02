package handlers

import (
	"github.com/JimcostDev/finances-api/services"
	"github.com/gofiber/fiber/v2"
)

type CategoryHandler struct {
	service services.CategoryService
}

func NewCategoryHandler(s services.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: s}
}

func (h *CategoryHandler) GetCategories(c *fiber.Ctx) error {
	categories, err := h.service.GetCategories(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	resp := make([]fiber.Map, 0, len(categories))
	for _, cat := range categories {
		resp = append(resp, fiber.Map{
			"id":     cat.ID.Hex(),
			"nombre": cat.Nombre,
			"tipo":   cat.Tipo,
		})
	}
	return c.JSON(resp)
}

