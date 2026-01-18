package routes

import (
	"github.com/JimcostDev/finances-api/handlers"
	"github.com/JimcostDev/finances-api/middleware"
	"github.com/gofiber/fiber/v2"
)

func ReportRoutes(app *fiber.App, handler *handlers.ReportHandler) {
	api := app.Group("/api/reports", middleware.Protected())

	// 1. Balance General (Histórico de todos los tiempos)
	api.Get("/general-balance", handler.GetGeneralBalance)

	// 2. Reporte Anual
	api.Get("/annual", handler.GetAnnualReport)

	// 3. Filtros y Listados
	api.Get("/by-month", handler.GetReportsByMonth)
	api.Get("/", handler.GetReports)
	api.Post("/", handler.CreateReport)

	// Operaciones CRUD sobre un reporte específico
	api.Get("/:id", handler.GetReportByID)
	api.Put("/:id", handler.UpdateReport)
	api.Delete("/:id", handler.DeleteReport)

	// Endpoints para modificar ingresos y gastos dentro de un reporte
	api.Post("/:id/income", handler.AddIncome)
	api.Delete("/:id/income/:income_id", handler.RemoveIncome)
	api.Post("/:id/expense", handler.AddExpense)
	api.Delete("/:id/expense/:expense_id", handler.RemoveExpense)
}
