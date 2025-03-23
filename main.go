package main

import (
	"log"

	"github.com/JimcostDev/finances-api/config"
	"github.com/JimcostDev/finances-api/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	app := fiber.New()

	// Configura CORS para permitir solicitudes desde localhost:4321
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:4321, https://finances-ui-app.vercel.app",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Conectar a MongoDB
	config.ConnectDB()

	// Configurar las rutas
	routes.SetupRoutes(app)

	// Hola Mundo
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Hola Mundo"})
	})

	log.Fatal(app.Listen(":3000"))
}
