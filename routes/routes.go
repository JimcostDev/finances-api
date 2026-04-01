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

	userRepo := repositories.NewUserRepository(config.DB)
	authService := services.NewAuthService(userRepo)
	authHandler := handlers.NewAuthHandler(authService)
	reportRepo := repositories.NewReportRepository(config.DB)
	reportService := services.NewReportService(reportRepo, userRepo)
	reportHandler := handlers.NewReportHandler(reportService)

	userService := services.NewUserService(userRepo, reportRepo, dbClient)
	userHandler := handlers.NewUserHandler(userService)

	AuthRoutes(app, authHandler)
	ReportRoutes(app, reportHandler)
	UserRoutes(app, userHandler)
}
