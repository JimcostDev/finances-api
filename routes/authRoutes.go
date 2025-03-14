// routes/authRoutes.go
package routes

import (
	"github.com/JimcostDev/finances-api/controllers"

	"github.com/gofiber/fiber/v2"
)

func AuthRoutes(app *fiber.App) {
	api := app.Group("/api/")
	api.Post("/register", controllers.Register)
	api.Post("/login", controllers.Login)
}
