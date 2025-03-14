// @title API de Finanzas Personales
// @version 1.0
// @description API para gestión de finanzas personales
// @contact.name Soporte API
// @contact.url https://www.jimcostdev.com
// @contact.email jimcostdev@gmail.com
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @host localhost:3000
// @BasePath /api
package main

import (
	"log"

	"github.com/JimcostDev/finances-api/config"
	"github.com/JimcostDev/finances-api/routes"

	_ "github.com/JimcostDev/finances-api/docs"
	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

func main() {
	app := fiber.New()

	// Conectar a MongoDB
	config.ConnectDB()

	// Configurar las rutas
	routes.SetupRoutes(app)

	// Ruta para Swagger
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	log.Fatal(app.Listen(":3000"))
}
