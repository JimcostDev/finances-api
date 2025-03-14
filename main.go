package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/JimcostDev/finances-api/config"
	"github.com/JimcostDev/finances-api/routes"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	// Conectar a MongoDB
	config.ConnectDB()

	// Configurar las rutas
	routes.SetupRoutes(app)

	// Hola Mundo
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Hola Mundo"})
	})

	// Endpoint para mostrar la lista de rutas
	app.Get("api/routes", func(c *fiber.Ctx) error {
		html := generateRoutesHTML(app.GetRoutes())
		c.Set("Content-Type", "text/html")
		return c.SendString(html)
	})

	log.Fatal(app.Listen(":3000"))
}

func generateRoutesHTML(routes []fiber.Route) string {
	var html strings.Builder
	html.WriteString("<h1>Lista de Endpoints</h1>")
	html.WriteString("<ul>")

	for _, route := range routes {
		html.WriteString(fmt.Sprintf("<li>%s %s</li>", route.Method, route.Path))
	}

	html.WriteString("</ul>")
	return html.String()
}
