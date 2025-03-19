package routes

import (
	"github.com/JimcostDev/finances-api/handlers"
	"github.com/JimcostDev/finances-api/middleware"
	"github.com/gofiber/fiber/v2"
)

func UserRoutes(app *fiber.App) {
	api := app.Group("/api/users/", middleware.Protected())
	api.Get("/profile", handlers.GetUserProfile)
	api.Patch("/profile", handlers.UpdateUser)
	api.Delete("/profile", handlers.DeleteUser)
}
