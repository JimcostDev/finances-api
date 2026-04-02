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
	reportRepo := repositories.NewReportRepository(config.DB)
	reportService := services.NewReportService(reportRepo, userRepo)
	userService := services.NewUserService(userRepo, reportRepo, dbClient, reportService)
	authHandler := handlers.NewAuthHandler(authService, userService)
	reportHandler := handlers.NewReportHandler(reportService)
	categoryRepo := repositories.NewCategoryRepository(config.DB)
	categoryService := services.NewCategoryService(categoryRepo)
	categoryHandler := handlers.NewCategoryHandler(categoryService)

	userHandler := handlers.NewUserHandler(userService)

	AuthRoutes(app, authHandler)
	ReportRoutes(app, reportHandler)
	CategoryRoutes(app, categoryHandler)
	UserRoutes(app, userHandler)
}
