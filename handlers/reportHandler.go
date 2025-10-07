package handlers

import (
	"context"
	"math"
	"strconv"
	"time"

	"github.com/JimcostDev/finances-api/config"
	"github.com/JimcostDev/finances-api/models"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Helper function para redondear a 2 decimales
func roundToTwoDecimals(value float64) float64 {
	return math.Round(value*100) / 100
}

// recalcReportTotals recalcula los totales del reporte basándose en sus ingresos y gastos.
func recalcReportTotals(report *models.Report) {
	totalIngresoBruto := 0.0
	for _, inc := range report.Ingresos {
		totalIngresoBruto += inc.Monto
	}

	totalGastos := 0.0
	for _, exp := range report.Gastos {
		totalGastos += exp.Monto
	}

	// Aplicar redondeo a todos los campos numéricos
	report.TotalIngresoBruto = roundToTwoDecimals(totalIngresoBruto)
	report.Diezmos = roundToTwoDecimals(totalIngresoBruto * 0.1)
	report.Ofrendas = roundToTwoDecimals(totalIngresoBruto * report.PorcentajeOfrenda)
	report.Iglesia = roundToTwoDecimals(report.Diezmos + report.Ofrendas)
	report.IngresosNetos = roundToTwoDecimals(totalIngresoBruto - report.Iglesia)
	report.TotalGastos = roundToTwoDecimals(totalGastos)
	report.Liquidacion = roundToTwoDecimals(report.IngresosNetos - report.TotalGastos)
	report.UpdatedAt = time.Now()
}

// CreateReport crea un nuevo reporte aplicando la lógica de negocio.
// @Summary Crea un nuevo reporte financiero
func CreateReport(c *fiber.Ctx) error {
	type ReportRequest struct {
		Month             string           `json:"month"`
		Year              int              `json:"year"`
		Ingresos          []models.Income  `json:"ingresos"`
		Gastos            []models.Expense `json:"gastos"`
		PorcentajeOfrenda float64          `json:"porcentaje_ofrenda"`
	}

	var req ReportRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"error": "Error al parsear JSON"})
	}

	// Redondear montos individuales y asignar un nuevo ObjectID a cada ingreso y gasto
	for i := range req.Ingresos {
		req.Ingresos[i].ID = primitive.NewObjectID()
		req.Ingresos[i].Monto = roundToTwoDecimals(req.Ingresos[i].Monto)
	}
	for i := range req.Gastos {
		req.Gastos[i].ID = primitive.NewObjectID()
		req.Gastos[i].Monto = roundToTwoDecimals(req.Gastos[i].Monto)
	}

	// Calcular totales con redondeo
	totalIngresoBruto := 0.0
	for _, inc := range req.Ingresos {
		totalIngresoBruto += inc.Monto
	}
	totalIngresoBruto = roundToTwoDecimals(totalIngresoBruto)

	totalGastos := 0.0
	for _, exp := range req.Gastos {
		totalGastos += exp.Monto
	}
	totalGastos = roundToTwoDecimals(totalGastos)

	// Cálculos con redondeo en cada paso
	diezmo := roundToTwoDecimals(totalIngresoBruto * 0.1)
	ofrenda := roundToTwoDecimals(totalIngresoBruto * req.PorcentajeOfrenda)
	iglesia := roundToTwoDecimals(diezmo + ofrenda)
	ingresosNetos := roundToTwoDecimals(totalIngresoBruto - iglesia)
	liquidacion := roundToTwoDecimals(ingresosNetos - totalGastos)

	// Obtener y validar userID
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).
			JSON(fiber.Map{"error": "Usuario no autenticado"})
	}

	userObjID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"error": "ID de usuario inválido"})
	}

	// Crear reporte con valores redondeados
	report := models.Report{
		UserID:            userObjID,
		Month:             req.Month,
		Year:              req.Year,
		Ingresos:          req.Ingresos,
		Gastos:            req.Gastos,
		PorcentajeOfrenda: req.PorcentajeOfrenda,
		TotalIngresoBruto: totalIngresoBruto,
		Diezmos:           diezmo,
		Ofrendas:          ofrenda,
		Iglesia:           iglesia,
		IngresosNetos:     ingresosNetos,
		TotalGastos:       totalGastos,
		Liquidacion:       liquidacion,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Insertar en la base de datos
	collection := config.DB.Collection("reports")
	result, err := collection.InsertOne(context.Background(), report)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"error": "Error al crear el reporte: " + err.Error()})
	}

	// Preparar respuesta
	response := fiber.Map{
		"id":                  result.InsertedID.(primitive.ObjectID).Hex(),
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

// UpdateReport actualiza un reporte existente y recalcula los campos según la lógica de negocio.
// @Summary Actualiza un reporte financiero existente
func UpdateReport(c *fiber.Ctx) error {
	id := c.Params("id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"error": "ID de reporte inválido"})
	}

	type ReportRequest struct {
		Month             string           `json:"month"`
		Year              int              `json:"year"`
		Ingresos          []models.Income  `json:"ingresos"`
		Gastos            []models.Expense `json:"gastos"`
		PorcentajeOfrenda float64          `json:"porcentaje_ofrenda"`
	}

	var req ReportRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"error": "Error al parsear JSON: " + err.Error()})
	}

	// Asignar nuevos IDs solo a elementos sin ID y redondear montos
	for i := range req.Ingresos {
		if req.Ingresos[i].ID.IsZero() {
			req.Ingresos[i].ID = primitive.NewObjectID()
		}
		req.Ingresos[i].Monto = roundToTwoDecimals(req.Ingresos[i].Monto)
	}

	for i := range req.Gastos {
		if req.Gastos[i].ID.IsZero() {
			req.Gastos[i].ID = primitive.NewObjectID()
		}
		req.Gastos[i].Monto = roundToTwoDecimals(req.Gastos[i].Monto)
	}

	// Calcular totales con redondeo
	totalIngresoBruto := roundToTwoDecimals(sumIngresos(req.Ingresos))
	totalGastos := roundToTwoDecimals(sumGastos(req.Gastos))

	// Cálculos con redondeo en cada paso
	diezmo := roundToTwoDecimals(totalIngresoBruto * 0.1)
	ofrenda := roundToTwoDecimals(totalIngresoBruto * req.PorcentajeOfrenda)
	iglesia := roundToTwoDecimals(diezmo + ofrenda)
	ingresosNetos := roundToTwoDecimals(totalIngresoBruto - iglesia)
	liquidacion := roundToTwoDecimals(ingresosNetos - totalGastos)

	// Validar usuario
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).
			JSON(fiber.Map{"error": "Usuario no autenticado"})
	}

	userObjID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"error": "ID de usuario inválido"})
	}

	// Construir actualización
	update := bson.M{
		"$set": bson.M{
			"month":               req.Month,
			"year":                req.Year,
			"ingresos":            req.Ingresos,
			"gastos":              req.Gastos,
			"porcentaje_ofrenda":  roundToTwoDecimals(req.PorcentajeOfrenda),
			"total_ingreso_bruto": totalIngresoBruto,
			"diezmos":             diezmo,
			"ofrendas":            ofrenda,
			"iglesia":             iglesia,
			"ingresos_netos":      ingresosNetos,
			"total_gastos":        totalGastos,
			"liquidacion":         liquidacion,
			"updated_at":          time.Now(),
		},
	}

	// Ejecutar actualización
	collection := config.DB.Collection("reports")
	result, err := collection.UpdateOne(
		context.Background(),
		bson.M{"_id": oid, "user_id": userObjID},
		update,
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"error": "Error al actualizar: " + err.Error()})
	}

	if result.MatchedCount == 0 {
		return c.Status(fiber.StatusNotFound).
			JSON(fiber.Map{"error": "Reporte no encontrado o no autorizado"})
	}

	return c.JSON(fiber.Map{
		"message": "Reporte actualizado exitosamente",
		"data": fiber.Map{
			"total_ingreso_bruto": totalIngresoBruto,
			"liquidacion":         liquidacion,
			"updated_at":          time.Now().Format(time.RFC3339),
		},
	})
}

// Funciones helper
func sumIngresos(ingresos []models.Income) float64 {
	total := 0.0
	for _, inc := range ingresos {
		total += inc.Monto
	}
	return total
}

func sumGastos(gastos []models.Expense) float64 {
	total := 0.0
	for _, exp := range gastos {
		total += exp.Monto
	}
	return total
}

// GetReports obtiene todos los reportes del usuario autenticado.
// @Summary Obtiene todos los reportes financieros del usuario autenticado
func GetReports(c *fiber.Ctx) error {
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).
			JSON(fiber.Map{"error": "Usuario no autenticado"})
	}
	userObjID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"error": "ID de usuario inválido"})
	}
	collection := config.DB.Collection("reports")
	filter := bson.M{"user_id": userObjID}

	// Opciones de búsqueda con ordenamiento
	opts := options.Find().SetSort(bson.D{
		{Key: "created_at", Value: -1},  // descendente
	})

	cursor, err := collection.Find(context.Background(), filter, opts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"error": "Error al obtener los reportes"})
	}
	defer cursor.Close(context.Background())
	var reports []models.Report
	for cursor.Next(context.Background()) {
		var report models.Report
		if err := cursor.Decode(&report); err != nil {
			return c.Status(fiber.StatusInternalServerError).
				JSON(fiber.Map{"error": "Error al decodificar reporte"})
		}
		reports = append(reports, report)
	}
	if err := cursor.Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"error": "Error en el cursor de reportes"})
	}
	// Convertir ObjectID a string para la respuesta
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

// GetReportByID obtiene un reporte específico por su ID, validando que pertenezca al usuario autenticado.
// @Summary Obtiene un reporte financiero por su ID
func GetReportByID(c *fiber.Ctx) error {
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).
			JSON(fiber.Map{"error": "Usuario no autenticado"})
	}
	userObjID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"error": "ID de usuario inválido"})
	}

	id := c.Params("id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"error": "ID inválido"})
	}

	collection := config.DB.Collection("reports")
	filter := bson.M{"_id": oid, "user_id": userObjID}

	var report models.Report
	err = collection.FindOne(context.Background(), filter).Decode(&report)
	if err != nil {
		return c.Status(fiber.StatusNotFound).
			JSON(fiber.Map{"error": "Reporte no encontrado"})
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

// GetReportsByMonth obtiene los reportes filtrados por mes y año.
// @Summary Obtiene reportes financieros filtrados por mes y año
func GetReportsByMonth(c *fiber.Ctx) error {
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).
			JSON(fiber.Map{"error": "Usuario no autenticado"})
	}
	userObjID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"error": "ID de usuario inválido"})
	}

	// Obtener parámetros de consulta (query params)
	month := c.Query("month")
	year, err := strconv.Atoi(c.Query("year"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"error": "El año debe ser un número válido"})
	}

	collection := config.DB.Collection("reports")
	filter := bson.M{"user_id": userObjID, "month": month, "year": year}

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"error": "Error al obtener los reportes"})
	}
	defer cursor.Close(context.Background())

	var reports []models.Report
	for cursor.Next(context.Background()) {
		var report models.Report
		if err := cursor.Decode(&report); err != nil {
			return c.Status(fiber.StatusInternalServerError).
				JSON(fiber.Map{"error": "Error al decodificar reporte"})
		}
		reports = append(reports, report)
	}

	if len(reports) == 0 {
		return c.JSON(fiber.Map{"message": "No se encontraron reportes para el mes y año especificados"})
	}

	return c.JSON(reports)
}

// DeleteReport elimina un reporte por su ID, validando que pertenezca al usuario autenticado.
// @Summary Elimina un reporte específico
func DeleteReport(c *fiber.Ctx) error {
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).
			JSON(fiber.Map{"error": "Usuario no autenticado"})
	}
	userObjID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"error": "ID de usuario inválido"})
	}

	id := c.Params("id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"error": "ID inválido"})
	}

	collection := config.DB.Collection("reports")
	filter := bson.M{"_id": oid, "user_id": userObjID}

	result, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"error": "Error al eliminar el reporte"})
	}
	if result.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).
			JSON(fiber.Map{"error": "Reporte no encontrado"})
	}

	return c.JSON(fiber.Map{"message": "Reporte eliminado exitosamente"})
}

// AddIncome agrega un nuevo ingreso a un reporte existente y recalcula los totales.
// @Summary Agrega un ingreso a un reporte
func AddIncome(c *fiber.Ctx) error {
	reportID := c.Params("id")
	oid, err := primitive.ObjectIDFromHex(reportID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID de reporte inválido"})
	}

	var newIncome models.Income
	if err := c.BodyParser(&newIncome); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Error al parsear JSON"})
	}
	// Asignar un nuevo ObjectID si no viene establecido o es cero.
	if newIncome.ID.IsZero() {
		newIncome.ID = primitive.NewObjectID()
	}

	// Obtener el userID desde el contexto.
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Usuario no autenticado"})
	}
	// Convertir el userID (suponemos que en Report.UserID se guarda como ObjectID).
	userObjID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ID de usuario inválido"})
	}

	collection := config.DB.Collection("reports")
	filter := bson.M{"_id": oid, "user_id": userObjID}

	// Buscar el reporte existente.
	var report models.Report
	if err := collection.FindOne(context.Background(), filter).Decode(&report); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Reporte no encontrado"})
	}

	// Agregar el nuevo ingreso.
	report.Ingresos = append(report.Ingresos, newIncome)
	// Recalcular totales.
	recalcReportTotals(&report)

	// Actualizar el documento completo.
	update := bson.M{"$set": report}
	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error al actualizar el reporte"})
	}

	return c.JSON(fiber.Map{"message": "Ingreso agregado exitosamente", "report": report})
}

// AddExpense agrega un nuevo gasto a un reporte existente y recalcula los totales.
// @Summary Agrega un gasto a un reporte
func AddExpense(c *fiber.Ctx) error {
	reportID := c.Params("id")
	oid, err := primitive.ObjectIDFromHex(reportID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID de reporte inválido"})
	}

	var newExpense models.Expense
	if err := c.BodyParser(&newExpense); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Error al parsear JSON"})
	}
	// Asignar un nuevo ObjectID si no viene establecido.
	if newExpense.ID.IsZero() {
		newExpense.ID = primitive.NewObjectID()
	}

	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Usuario no autenticado"})
	}
	userObjID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ID de usuario inválido"})
	}

	collection := config.DB.Collection("reports")
	filter := bson.M{"_id": oid, "user_id": userObjID}

	var report models.Report
	if err := collection.FindOne(context.Background(), filter).Decode(&report); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Reporte no encontrado"})
	}

	// Agregar el nuevo gasto.
	report.Gastos = append(report.Gastos, newExpense)
	// Recalcular totales.
	recalcReportTotals(&report)

	update := bson.M{"$set": report}
	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error al actualizar el reporte"})
	}

	return c.JSON(fiber.Map{"message": "Gasto agregado exitosamente", "report": report})
}

// RemoveIncome elimina un ingreso específico de un reporte y recalcula los totales.
// @Summary Elimina un ingreso de un reporte
func RemoveIncome(c *fiber.Ctx) error {
	reportID := c.Params("id")
	incomeID := c.Params("income_id")

	oid, err := primitive.ObjectIDFromHex(reportID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID de reporte inválido"})
	}

	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Usuario no autenticado"})
	}
	userObjID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ID de usuario inválido"})
	}

	collection := config.DB.Collection("reports")
	filter := bson.M{"_id": oid, "user_id": userObjID}

	var report models.Report
	if err := collection.FindOne(context.Background(), filter).Decode(&report); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Reporte no encontrado"})
	}

	// Filtrar el ingreso a eliminar.
	var updatedIncomes []models.Income
	found := false
	for _, inc := range report.Ingresos {
		if inc.ID.Hex() != incomeID {
			updatedIncomes = append(updatedIncomes, inc)
		} else {
			found = true
		}
	}
	if !found {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Ingreso no encontrado"})
	}

	report.Ingresos = updatedIncomes
	recalcReportTotals(&report)

	update := bson.M{"$set": report}
	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error al actualizar el reporte"})
	}

	return c.JSON(fiber.Map{"message": "Ingreso eliminado exitosamente", "report": report})
}

// RemoveExpense elimina un gasto específico de un reporte y recalcula los totales.
// @Summary Elimina un gasto de un reporte
func RemoveExpense(c *fiber.Ctx) error {
	reportID := c.Params("id")
	expenseID := c.Params("expense_id")

	oid, err := primitive.ObjectIDFromHex(reportID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID de reporte inválido"})
	}

	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Usuario no autenticado"})
	}
	userObjID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ID de usuario inválido"})
	}

	collection := config.DB.Collection("reports")
	filter := bson.M{"_id": oid, "user_id": userObjID}

	var report models.Report
	if err := collection.FindOne(context.Background(), filter).Decode(&report); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Reporte no encontrado"})
	}

	var updatedExpenses []models.Expense
	found := false
	for _, exp := range report.Gastos {
		if exp.ID.Hex() != expenseID {
			updatedExpenses = append(updatedExpenses, exp)
		} else {
			found = true
		}
	}
	if !found {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Gasto no encontrado"})
	}

	report.Gastos = updatedExpenses
	recalcReportTotals(&report)

	update := bson.M{"$set": report}
	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error al actualizar el reporte"})
	}

	return c.JSON(fiber.Map{"message": "Gasto eliminado exitosamente", "report": report})
}

// GetAnnualReport obtiene el reporte anual para el usuario autenticado,
// sumando los totales de todos los reportes de un año específico.
func GetAnnualReport(c *fiber.Ctx) error {
	// Extraer el año de los query params
	yearStr := c.Query("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El año debe ser un número válido"})
	}

	// Obtener el userID desde el contexto
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Usuario no autenticado"})
	}
	userObjID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID de usuario inválido"})
	}

	// Definir el pipeline de agregación
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{
			{Key: "year", Value: year},
			{Key: "user_id", Value: userObjID},
		}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "total_ingreso_bruto", Value: bson.D{{Key: "$sum", Value: "$total_ingreso_bruto"}}},
			{Key: "total_ingreso_neto", Value: bson.D{{Key: "$sum", Value: "$ingresos_netos"}}},
			{Key: "total_diezmos", Value: bson.D{{Key: "$sum", Value: "$diezmos"}}},
			{Key: "total_ofrendas", Value: bson.D{{Key: "$sum", Value: "$ofrendas"}}},
			{Key: "total_iglesia", Value: bson.D{{Key: "$sum", Value: "$iglesia"}}},
			{Key: "total_gastos", Value: bson.D{{Key: "$sum", Value: "$total_gastos"}}},
			{Key: "liquidacion_final", Value: bson.D{{Key: "$sum", Value: "$liquidacion"}}},
			{Key: "user_id", Value: bson.D{{Key: "$first", Value: "$user_id"}}},
			{Key: "year", Value: bson.D{{Key: "$first", Value: "$year"}}},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "user_id", Value: 1},
			{Key: "year", Value: 1},
			{Key: "total_ingreso_bruto", Value: 1},
			{Key: "total_ingreso_neto", Value: 1},
			{Key: "total_diezmos", Value: 1},
			{Key: "total_ofrendas", Value: 1},
			{Key: "total_iglesia", Value: 1},
			{Key: "total_gastos", Value: 1},
			{Key: "liquidacion_final", Value: 1},
		}}},
	}

	collection := config.DB.Collection("reports")
	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error al obtener el reporte anual"})
	}
	var results []bson.M
	if err = cursor.All(context.Background(), &results); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error al decodificar el reporte anual"})
	}

	if len(results) == 0 {
		return c.JSON(fiber.Map{"message": "No se encontraron reportes para el año especificado"})
	}

	// Redondear cada valor a 2 decimales usando el helper
	result := results[0]
	if val, ok := result["total_ingreso_bruto"].(float64); ok {
		result["total_ingreso_bruto"] = roundToTwoDecimals(val)
	}
	if val, ok := result["total_ingreso_neto"].(float64); ok {
		result["total_ingreso_neto"] = roundToTwoDecimals(val)
	}
	if val, ok := result["total_diezmos"].(float64); ok {
		result["total_diezmos"] = roundToTwoDecimals(val)
	}
	if val, ok := result["total_ofrendas"].(float64); ok {
		result["total_ofrendas"] = roundToTwoDecimals(val)
	}
	if val, ok := result["total_iglesia"].(float64); ok {
		result["total_iglesia"] = roundToTwoDecimals(val)
	}
	if val, ok := result["total_gastos"].(float64); ok {
		result["total_gastos"] = roundToTwoDecimals(val)
	}
	if val, ok := result["liquidacion_final"].(float64); ok {
		result["liquidacion_final"] = roundToTwoDecimals(val)
	}

	return c.JSON(results[0])
}
