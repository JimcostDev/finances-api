package handlers

import (
	"context"
	"strconv"
	"time"

	"github.com/JimcostDev/finances-api/config"
	"github.com/JimcostDev/finances-api/models"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

	report.TotalIngresoBruto = totalIngresoBruto
	report.Diezmos = totalIngresoBruto * 0.1
	report.Ofrendas = totalIngresoBruto * report.PorcentajeOfrenda
	report.Iglesia = report.Diezmos + report.Ofrendas
	report.IngresosNetos = totalIngresoBruto - report.Iglesia
	report.TotalGastos = totalGastos
	report.Liquidacion = report.IngresosNetos - totalGastos
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
		PorcentajeOfrenda float64          `json:"porcentaje_ofrenda"` // Ej: 0.04 para 4%
	}
	var req ReportRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"error": "Error al parsear JSON"})
	}

	// Asignar un nuevo ObjectID a cada ingreso y gasto
	for i := range req.Ingresos {
		req.Ingresos[i].ID = primitive.NewObjectID()
	}
	for i := range req.Gastos {
		req.Gastos[i].ID = primitive.NewObjectID()
	}

	// Calcular totales.
	totalIngresoBruto := 0.0
	for _, inc := range req.Ingresos {
		totalIngresoBruto += inc.Monto
	}
	totalGastos := 0.0
	for _, exp := range req.Gastos {
		totalGastos += exp.Monto
	}

	// Lógica de negocio.
	diezmo := totalIngresoBruto * 0.1
	ofrenda := totalIngresoBruto * req.PorcentajeOfrenda
	iglesia := diezmo + ofrenda
	ingresosNetos := totalIngresoBruto - iglesia
	liquidacion := ingresosNetos - totalGastos

	// Obtener el userID desde el contexto y convertirlo a ObjectID.
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

	collection := config.DB.Collection("reports")
	result, err := collection.InsertOne(context.Background(), report)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"error": "Error al insertar el reporte"})
	}

	// Asignar el ObjectID generado.
	report.ID = result.InsertedID.(primitive.ObjectID)

	// Devolver la respuesta convirtiendo los ObjectID a string (hex).
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
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

// UpdateReport actualiza un reporte existente y recalcula los campos según la lógica de negocio.
// @Summary Actualiza un reporte financiero existente
func UpdateReport(c *fiber.Ctx) error {
	id := c.Params("id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"error": "ID inválido"})
	}

	type ReportRequest struct {
		Month             string           `json:"month"`
		Year              int              `json:"year"`
		Ingresos          []models.Income  `json:"ingresos"`
		Gastos            []models.Expense `json:"gastos"`
		PorcentajeOfrenda float64          `json:"porcentaje_ofrenda"` // Ej: 0.04 para 4%
	}
	var req ReportRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"error": "Error al parsear JSON"})
	}

	// Asignar un nuevo ObjectID a cada ingreso y gasto
	for i := range req.Ingresos {
		req.Ingresos[i].ID = primitive.NewObjectID()
	}
	for i := range req.Gastos {
		req.Gastos[i].ID = primitive.NewObjectID()
	}

	totalIngresoBruto := 0.0
	for _, inc := range req.Ingresos {
		totalIngresoBruto += inc.Monto
	}
	totalGastos := 0.0
	for _, exp := range req.Gastos {
		totalGastos += exp.Monto
	}
	diezmo := totalIngresoBruto * 0.1
	ofrenda := totalIngresoBruto * req.PorcentajeOfrenda
	iglesia := diezmo + ofrenda
	ingresosNetos := totalIngresoBruto - iglesia
	liquidacion := ingresosNetos - totalGastos

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

	update := bson.M{
		"$set": bson.M{
			"month":               req.Month,
			"year":                req.Year,
			"ingresos":            req.Ingresos,
			"gastos":              req.Gastos,
			"porcentaje_ofrenda":  req.PorcentajeOfrenda,
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

	collection := config.DB.Collection("reports")
	// Filtrar por _id y user_id (como ObjectID)
	_, err = collection.UpdateOne(context.Background(), bson.M{
		"_id":     oid,
		"user_id": userObjID,
	}, update)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"error": "Error al actualizar el reporte"})
	}

	return c.JSON(fiber.Map{"message": "Reporte actualizado exitosamente"})
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
