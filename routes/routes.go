package routes

import "github.com/gofiber/fiber/v2"

func SetupRoutes(app *fiber.App) {
	AuthRoutes(app)
	ReportRoutes(app)
	UserRoutes(app)
}
