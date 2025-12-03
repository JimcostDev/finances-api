package routes

import (
	"github.com/JimcostDev/finances-api/handlers"
	"github.com/JimcostDev/finances-api/middleware"
	"github.com/gofiber/fiber/v2"
)

// Modificamos la función para recibir el 'handler' como parámetro
func ReportRoutes(app *fiber.App, handler *handlers.ReportHandler) {
	api := app.Group("/api/reports", middleware.Protected())

	// Rutas básicas de Reporte usando los MÉTODOS del handler
	api.Get("/by-month", handler.GetReportsByMonth)
	api.Get("/", handler.GetReports)
	api.Get("/annual", handler.GetAnnualReport)
	api.Post("/", handler.CreateReport)
	api.Get("/:id", handler.GetReportByID)
	api.Put("/:id", handler.UpdateReport)
	api.Delete("/:id", handler.DeleteReport)

	// Endpoints para modificar ingresos y gastos
	api.Post("/:id/income", handler.AddIncome)
	api.Delete("/:id/income/:income_id", handler.RemoveIncome)
	api.Post("/:id/expense", handler.AddExpense)
	api.Delete("/:id/expense/:expense_id", handler.RemoveExpense)
}
