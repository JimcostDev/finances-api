package routes

import (
	"github.com/JimcostDev/finances-api/handlers"
	"github.com/JimcostDev/finances-api/middleware"
	"github.com/gofiber/fiber/v2"
)

func CategoryRoutes(app *fiber.App, handler *handlers.CategoryHandler) {
	api := app.Group("/api/categories", middleware.Protected())
	api.Get("/", handler.GetCategories)
}

