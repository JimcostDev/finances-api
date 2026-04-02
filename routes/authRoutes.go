package routes

import (
	"github.com/JimcostDev/finances-api/handlers"
	"github.com/JimcostDev/finances-api/middleware"
	"github.com/gofiber/fiber/v2"
)

// AuthRoutes ahora recibe el AuthHandler
func AuthRoutes(app *fiber.App, handler *handlers.AuthHandler) {
	api := app.Group("api/auth")

	api.Post("/register", handler.Register)
	api.Post("/login", handler.Login)
	api.Post("/logout", handler.Logout)
	api.Get("/me", middleware.Protected(), handler.Me)
}
