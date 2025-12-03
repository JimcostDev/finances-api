package routes

import (
	"github.com/JimcostDev/finances-api/handlers"
	"github.com/JimcostDev/finances-api/middleware"
	"github.com/gofiber/fiber/v2"
)

func UserRoutes(app *fiber.App, handler *handlers.UserHandler) {
	api := app.Group("/api/users", middleware.Protected())

	api.Get("/profile", handler.GetUserProfile)
	api.Put("/profile", handler.UpdateUser)
	api.Delete("/profile", handler.DeleteUser)
}
