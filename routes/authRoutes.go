// routes/authRoutes.go
package routes

import (
	"github.com/JimcostDev/finances-api/handlers"

	"github.com/gofiber/fiber/v2"
)

func AuthRoutes(app *fiber.App) {
	api := app.Group("/api/")
	api.Post("/register", handlers.Register)
	api.Post("/login", handlers.Login)
}
