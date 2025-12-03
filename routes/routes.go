package routes

import (
	"github.com/JimcostDev/finances-api/config"
	"github.com/JimcostDev/finances-api/handlers"
	"github.com/JimcostDev/finances-api/repositories"
	"github.com/JimcostDev/finances-api/services"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	// Obtenemos el cliente subyacente para manejar sesiones/transacciones
	dbClient := config.DB.Client()

	// --- AUTHENTICATION (Wiring) ---
	userRepo := repositories.NewUserRepository(config.DB)
	authService := services.NewAuthService(userRepo)
	authHandler := handlers.NewAuthHandler(authService)

	// --- REPORTS (Wiring) ---
	reportRepo := repositories.NewReportRepository(config.DB)
	reportService := services.NewReportService(reportRepo)
	reportHandler := handlers.NewReportHandler(reportService)

	// --- USER (Wiring) ---
	// El servicio de usuario necesita:
	// 1. userRepo (para editar/borrar usuario)
	// 2. reportRepo (para borrar reportes en cascada)
	// 3. dbClient (para ejecutar la transacción de borrado completo)
	userService := services.NewUserService(userRepo, reportRepo, dbClient)
	userHandler := handlers.NewUserHandler(userService)

	// --- REGISTRO DE RUTAS ---
	AuthRoutes(app, authHandler)
	ReportRoutes(app, reportHandler)
	UserRoutes(app, userHandler)
}
