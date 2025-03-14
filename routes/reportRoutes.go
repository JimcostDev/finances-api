package routes

import (
	"github.com/JimcostDev/finances-api/controllers"
	"github.com/JimcostDev/finances-api/middleware"
	"github.com/gofiber/fiber/v2"
)

func ReportRoutes(app *fiber.App) {
	api := app.Group("/api/reports", middleware.Protected())

	// Rutas básicas de Reporte
	api.Get("/by-month", controllers.GetReportsByMonth)
	api.Get("/", controllers.GetReports)
	api.Post("/", controllers.CreateReport)
	api.Get("/:id", controllers.GetReportByID)
	api.Put("/:id", controllers.UpdateReport)
	api.Delete("/:id", controllers.DeleteReport)

	// Endpoints para modificar ingresos y gastos dentro del reporte
	api.Post("/:id/income", controllers.AddIncome)
	api.Delete("/:id/income/:income_id", controllers.RemoveIncome)
	api.Post("/:id/expense", controllers.AddExpense)
	api.Delete("/:id/expense/:expense_id", controllers.RemoveExpense)
}
