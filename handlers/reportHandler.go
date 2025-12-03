package handlers

import (
	"strconv"

	"github.com/JimcostDev/finances-api/models"
	"github.com/JimcostDev/finances-api/services"
	"github.com/gofiber/fiber/v2"
)

type ReportHandler struct {
	service services.ReportService
}

func NewReportHandler(s services.ReportService) *ReportHandler {
	return &ReportHandler{service: s}
}

// CreateReport crea un nuevo reporte
func (h *ReportHandler) CreateReport(c *fiber.Ctx) error {
	var req services.ReportRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Error al parsear JSON"})
	}

	userID := c.Locals("userID").(string)
	report, err := h.service.CreateReport(c.Context(), userID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Mapeo manual para asegurar que la respuesta sea idéntica al frontend
	response := fiber.Map{
		"id":                  report.ID.Hex(),
		"user_id":             report.UserID.Hex(),
		"month":               report.Month,
		"year":                report.Year,
		"total_ingreso_bruto": report.TotalIngresoBruto,
		"diezmos":             report.Diezmos,
		"ofrendas":            report.Ofrendas,
		"iglesia":             report.Iglesia,
		"ingresos_netos":      report.IngresosNetos,
		"total_gastos":        report.TotalGastos,
		"liquidacion":         report.Liquidacion,
		"created_at":          report.CreatedAt,
		"updated_at":          report.UpdatedAt,
	}
	return c.Status(fiber.StatusCreated).JSON(response)
}

// UpdateReport actualiza un reporte
func (h *ReportHandler) UpdateReport(c *fiber.Ctx) error {
	id := c.Params("id")
	var req services.ReportRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Error al parsear JSON"})
	}

	userID := c.Locals("userID").(string)
	result, err := h.service.UpdateReport(c.Context(), id, userID, req)
	if err != nil {
		if err.Error() == "not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Reporte no encontrado"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Reporte actualizado exitosamente",
		"data":    result,
	})
}

// GetReports obtiene todos
func (h *ReportHandler) GetReports(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	reports, err := h.service.GetReports(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Conversión para el frontend (ObjectID a Hex string)
	var reportsResp []fiber.Map
	for _, report := range reports {
		reportsResp = append(reportsResp, fiber.Map{
			"id":                  report.ID.Hex(),
			"user_id":             report.UserID.Hex(),
			"month":               report.Month,
			"year":                report.Year,
			"ingresos":            report.Ingresos,
			"gastos":              report.Gastos,
			"porcentaje_ofrenda":  report.PorcentajeOfrenda,
			"total_ingreso_bruto": report.TotalIngresoBruto,
			"diezmos":             report.Diezmos,
			"ofrendas":            report.Ofrendas,
			"iglesia":             report.Iglesia,
			"ingresos_netos":      report.IngresosNetos,
			"total_gastos":        report.TotalGastos,
			"liquidacion":         report.Liquidacion,
			"created_at":          report.CreatedAt,
			"updated_at":          report.UpdatedAt,
		})
	}
	return c.JSON(reportsResp)
}

// GetReportByID un reporte
func (h *ReportHandler) GetReportByID(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	id := c.Params("id")

	report, err := h.service.GetReportByID(c.Context(), id, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Reporte no encontrado"})
	}

	return c.JSON(fiber.Map{
		"id":                  report.ID.Hex(),
		"user_id":             report.UserID.Hex(),
		"month":               report.Month,
		"year":                report.Year,
		"ingresos":            report.Ingresos,
		"gastos":              report.Gastos,
		"porcentaje_ofrenda":  report.PorcentajeOfrenda,
		"total_ingreso_bruto": report.TotalIngresoBruto,
		"diezmos":             report.Diezmos,
		"ofrendas":            report.Ofrendas,
		"iglesia":             report.Iglesia,
		"ingresos_netos":      report.IngresosNetos,
		"total_gastos":        report.TotalGastos,
		"liquidacion":         report.Liquidacion,
		"created_at":          report.CreatedAt,
		"updated_at":          report.UpdatedAt,
	})
}

// GetReportsByMonth filtro
func (h *ReportHandler) GetReportsByMonth(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	month := c.Query("month")
	year, err := strconv.Atoi(c.Query("year"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El año debe ser válido"})
	}

	reports, err := h.service.GetReportsByMonth(c.Context(), userID, month, year)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if len(reports) == 0 {
		return c.JSON(fiber.Map{"message": "No se encontraron reportes"})
	}
	return c.JSON(reports)
}

// DeleteReport elimina
func (h *ReportHandler) DeleteReport(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	id := c.Params("id")

	err := h.service.DeleteReport(c.Context(), id, userID)
	if err != nil {
		if err.Error() == "not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Reporte no encontrado"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Reporte eliminado exitosamente"})
}

// AddIncome
func (h *ReportHandler) AddIncome(c *fiber.Ctx) error {
	reportID := c.Params("id")
	var newIncome models.Income
	if err := c.BodyParser(&newIncome); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Error al parsear JSON"})
	}

	userID := c.Locals("userID").(string)
	report, err := h.service.AddIncome(c.Context(), reportID, userID, newIncome)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Ingreso agregado exitosamente", "report": report})
}

// AddExpense
func (h *ReportHandler) AddExpense(c *fiber.Ctx) error {
	reportID := c.Params("id")
	var newExpense models.Expense
	if err := c.BodyParser(&newExpense); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Error al parsear JSON"})
	}

	userID := c.Locals("userID").(string)
	report, err := h.service.AddExpense(c.Context(), reportID, userID, newExpense)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Gasto agregado exitosamente", "report": report})
}

// RemoveIncome
func (h *ReportHandler) RemoveIncome(c *fiber.Ctx) error {
	reportID := c.Params("id")
	incomeID := c.Params("income_id")
	userID := c.Locals("userID").(string)

	report, err := h.service.RemoveIncome(c.Context(), reportID, userID, incomeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Ingreso eliminado exitosamente", "report": report})
}

// RemoveExpense
func (h *ReportHandler) RemoveExpense(c *fiber.Ctx) error {
	reportID := c.Params("id")
	expenseID := c.Params("expense_id")
	userID := c.Locals("userID").(string)

	report, err := h.service.RemoveExpense(c.Context(), reportID, userID, expenseID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Gasto eliminado exitosamente", "report": report})
}

// GetAnnualReport
func (h *ReportHandler) GetAnnualReport(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	yearStr := c.Query("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El año debe ser válido"})
	}

	result, err := h.service.GetAnnualReport(c.Context(), userID, year)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if result == nil {
		return c.JSON(fiber.Map{"message": "No se encontraron reportes para el año especificado"})
	}
	return c.JSON(result)
}
