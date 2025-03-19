package routes

import (
	"github.com/JimcostDev/finances-api/handlers"
	"github.com/JimcostDev/finances-api/middleware"
	"github.com/gofiber/fiber/v2"
)

func UserRoutes(app *fiber.App) {
	api := app.Group("/api/users/", middleware.Protected())
	api.Get("/:id", handlers.GetUser)
	api.Patch("/:id", handlers.UpdateUser)
	api.Delete("/:id", handlers.DeleteUser)
}
