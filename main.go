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

	// Configura CORS para permitir solicitudes desde localhost:4321 y el dominio de producción
	app.Use(cors.New(cors.Config{
		AllowOrigins: "https://finances.jimcostdev.com, http://localhost:4321",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Conectar a MongoDB. (Paso crucial antes de cargar las rutas)
	config.ConnectDB()

	// Configurar las rutas. (Esto inicia la inyección de dependencias)
	routes.SetupRoutes(app)

	// Hola Mundo / Health Check
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Hola Mundo"})
	})

	log.Fatal(app.Listen(":3000"))
}
