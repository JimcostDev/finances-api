package routes

import (
	"github.com/JimcostDev/finances-api/handlers"
	"github.com/JimcostDev/finances-api/middleware"
	"github.com/gofiber/fiber/v2"
)

func ReportRoutes(app *fiber.App) {
	api := app.Group("/api/reports", middleware.Protected())

	// Rutas básicas de Reporte
	api.Get("/by-month", handlers.GetReportsByMonth)
	api.Get("/", handlers.GetReports)
	api.Get("/annual", handlers.GetAnnualReport)
	api.Post("/", handlers.CreateReport)
	api.Get("/:id", handlers.GetReportByID)
	api.Put("/:id", handlers.UpdateReport)
	api.Delete("/:id", handlers.DeleteReport)

	// Endpoints para modificar ingresos y gastos dentro del reporte
	api.Post("/:id/income", handlers.AddIncome)
	api.Delete("/:id/income/:income_id", handlers.RemoveIncome)
	api.Post("/:id/expense", handlers.AddExpense)
	api.Delete("/:id/expense/:expense_id", handlers.RemoveExpense)
}
