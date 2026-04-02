package main

import (
	"log"
	"os"
	"strings"

	"github.com/JimcostDev/finances-api/config"
	"github.com/JimcostDev/finances-api/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	// Tras proxy (p. ej. Koyeb), c.Protocol() y cookies Secure usan X-Forwarded-Proto
	app := fiber.New(fiber.Config{
		ProxyHeader: "X-Forwarded-Proto",
	})

	// CORS: credenciales necesarias para cookies cross-origin (añade orígenes en CORS_ORIGINS separados por coma)
	allowOrigins := os.Getenv("CORS_ORIGINS")
	if allowOrigins == "" {
		allowOrigins = "https://finances.jimcostdev.com,http://localhost:4321,http://localhost:3000,http://127.0.0.1:4321"
	}
	parts := strings.Split(allowOrigins, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Join(parts, ","),
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))

	// Conectar a MongoDB.
	config.ConnectDB()

	// Configurar las rutas
	routes.SetupRoutes(app)

	// Hola Mundo / Health Check
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Hola Mundo"})
	})

	log.Fatal(app.Listen(":3000"))
}
