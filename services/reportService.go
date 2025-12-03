package services

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/JimcostDev/finances-api/models"
	"github.com/JimcostDev/finances-api/repositories"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ReportService interface {
	CreateReport(ctx context.Context, userID string, req ReportRequest) (*models.Report, error)
	UpdateReport(ctx context.Context, reportID string, userID string, req ReportRequest) (interface{}, error)
	GetReports(ctx context.Context, userID string) ([]models.Report, error)
	GetReportByID(ctx context.Context, reportID string, userID string) (*models.Report, error)
	GetReportsByMonth(ctx context.Context, userID string, month string, year int) ([]models.Report, error)
	DeleteReport(ctx context.Context, reportID string, userID string) error
	GetAnnualReport(ctx context.Context, userID string, year int) (bson.M, error)

	// Métodos para items individuales
	AddIncome(ctx context.Context, reportID, userID string, income models.Income) (*models.Report, error)
	AddExpense(ctx context.Context, reportID, userID string, expense models.Expense) (*models.Report, error)
	RemoveIncome(ctx context.Context, reportID, userID, incomeID string) (*models.Report, error)
	RemoveExpense(ctx context.Context, reportID, userID, expenseID string) (*models.Report, error)
}

// Estructura auxiliar para recibir datos (la moví del handler aquí)
type ReportRequest struct {
	Month             string           `json:"month"`
	Year              int              `json:"year"`
	Ingresos          []models.Income  `json:"ingresos"`
	Gastos            []models.Expense `json:"gastos"`
	PorcentajeOfrenda float64          `json:"porcentaje_ofrenda"`
}

type reportService struct {
	repo repositories.ReportRepository
}

func NewReportService(repo repositories.ReportRepository) ReportService {
	return &reportService{repo: repo}
}

// --- Helpers de Lógica de Negocio ---
func roundToTwoDecimals(value float64) float64 {
	return math.Round(value*100) / 100
}

func recalcReportTotals(report *models.Report) {
	totalIngresoBruto := 0.0
	for _, inc := range report.Ingresos {
		totalIngresoBruto += inc.Monto
	}

	totalGastos := 0.0
	for _, exp := range report.Gastos {
		totalGastos += exp.Monto
	}

	report.TotalIngresoBruto = roundToTwoDecimals(totalIngresoBruto)
	report.Diezmos = roundToTwoDecimals(totalIngresoBruto * 0.1)
	report.Ofrendas = roundToTwoDecimals(totalIngresoBruto * report.PorcentajeOfrenda)
	report.Iglesia = roundToTwoDecimals(report.Diezmos + report.Ofrendas)
	report.IngresosNetos = roundToTwoDecimals(totalIngresoBruto - report.Iglesia)
	report.TotalGastos = roundToTwoDecimals(totalGastos)
	report.Liquidacion = roundToTwoDecimals(report.IngresosNetos - report.TotalGastos)
	report.UpdatedAt = time.Now()
}

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

// --- Implementación de Métodos ---

func (s *reportService) CreateReport(ctx context.Context, userIDStr string, req ReportRequest) (*models.Report, error) {
	userObjID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	// Lógica de IDs y Redondeo
	for i := range req.Ingresos {
		req.Ingresos[i].ID = primitive.NewObjectID()
		req.Ingresos[i].Monto = roundToTwoDecimals(req.Ingresos[i].Monto)
	}
	for i := range req.Gastos {
		req.Gastos[i].ID = primitive.NewObjectID()
		req.Gastos[i].Monto = roundToTwoDecimals(req.Gastos[i].Monto)
	}

	// Crear struct temporal para usar recalcReportTotals
	tempReport := models.Report{
		Ingresos:          req.Ingresos,
		Gastos:            req.Gastos,
		PorcentajeOfrenda: req.PorcentajeOfrenda,
	}
	recalcReportTotals(&tempReport)

	finalReport := models.Report{
		UserID:            userObjID,
		Month:             req.Month,
		Year:              req.Year,
		Ingresos:          req.Ingresos,
		Gastos:            req.Gastos,
		PorcentajeOfrenda: req.PorcentajeOfrenda,
		TotalIngresoBruto: tempReport.TotalIngresoBruto,
		Diezmos:           tempReport.Diezmos,
		Ofrendas:          tempReport.Ofrendas,
		Iglesia:           tempReport.Iglesia,
		IngresosNetos:     tempReport.IngresosNetos,
		TotalGastos:       tempReport.TotalGastos,
		Liquidacion:       tempReport.Liquidacion,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	res, err := s.repo.Create(ctx, finalReport)
	if err != nil {
		return nil, err
	}

	finalReport.ID = res.InsertedID.(primitive.ObjectID)
	return &finalReport, nil
}

func (s *reportService) UpdateReport(ctx context.Context, reportID string, userIDStr string, req ReportRequest) (interface{}, error) {
	oid, err := primitive.ObjectIDFromHex(reportID)
	if err != nil {
		return nil, errors.New("invalid report ID")
	}
	userObjID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	// Lógica de redondeo y asignación de IDs si faltan
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

	// Usamos la lógica de cálculo
	tempReport := models.Report{
		Ingresos:          req.Ingresos,
		Gastos:            req.Gastos,
		PorcentajeOfrenda: req.PorcentajeOfrenda,
	}
	recalcReportTotals(&tempReport)

	update := bson.M{
		"$set": bson.M{
			"month":               req.Month,
			"year":                req.Year,
			"ingresos":            req.Ingresos,
			"gastos":              req.Gastos,
			"porcentaje_ofrenda":  roundToTwoDecimals(req.PorcentajeOfrenda),
			"total_ingreso_bruto": tempReport.TotalIngresoBruto,
			"diezmos":             tempReport.Diezmos,
			"ofrendas":            tempReport.Ofrendas,
			"iglesia":             tempReport.Iglesia,
			"ingresos_netos":      tempReport.IngresosNetos,
			"total_gastos":        tempReport.TotalGastos,
			"liquidacion":         tempReport.Liquidacion,
			"updated_at":          time.Now(),
		},
	}

	result, err := s.repo.Update(ctx, oid, userObjID, update)
	if err != nil {
		return nil, err
	}
	if result.MatchedCount == 0 {
		return nil, errors.New("not found")
	}

	// Retornamos datos para la respuesta
	return map[string]interface{}{
		"total_ingreso_bruto": tempReport.TotalIngresoBruto,
		"liquidacion":         tempReport.Liquidacion,
		"updated_at":          time.Now().Format(time.RFC3339),
	}, nil
}

func (s *reportService) GetReports(ctx context.Context, userIDStr string) ([]models.Report, error) {
	userObjID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}
	return s.repo.FindAll(ctx, userObjID)
}

func (s *reportService) GetReportByID(ctx context.Context, reportID, userIDStr string) (*models.Report, error) {
	oid, err := primitive.ObjectIDFromHex(reportID)
	if err != nil {
		return nil, errors.New("invalid report ID")
	}
	userObjID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}
	return s.repo.FindOne(ctx, oid, userObjID)
}

func (s *reportService) GetReportsByMonth(ctx context.Context, userIDStr, month string, year int) ([]models.Report, error) {
	userObjID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}
	return s.repo.FindByMonth(ctx, userObjID, month, year)
}

func (s *reportService) DeleteReport(ctx context.Context, reportID, userIDStr string) error {
	oid, err := primitive.ObjectIDFromHex(reportID)
	if err != nil {
		return errors.New("invalid report ID")
	}
	userObjID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return errors.New("invalid user ID")
	}

	res, err := s.repo.Delete(ctx, oid, userObjID)
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.New("not found")
	}
	return nil
}

func (s *reportService) AddIncome(ctx context.Context, reportID, userIDStr string, newIncome models.Income) (*models.Report, error) {
	if newIncome.ID.IsZero() {
		newIncome.ID = primitive.NewObjectID()
	}

	// Reutilizamos GetReportByID
	report, err := s.GetReportByID(ctx, reportID, userIDStr)
	if err != nil {
		return nil, err
	}

	report.Ingresos = append(report.Ingresos, newIncome)
	recalcReportTotals(report)

	userObjID, _ := primitive.ObjectIDFromHex(userIDStr) // Ya validado arriba
	_, err = s.repo.Update(ctx, report.ID, userObjID, bson.M{"$set": report})
	return report, err
}

func (s *reportService) AddExpense(ctx context.Context, reportID, userIDStr string, newExpense models.Expense) (*models.Report, error) {
	if newExpense.ID.IsZero() {
		newExpense.ID = primitive.NewObjectID()
	}

	report, err := s.GetReportByID(ctx, reportID, userIDStr)
	if err != nil {
		return nil, err
	}

	report.Gastos = append(report.Gastos, newExpense)
	recalcReportTotals(report)

	userObjID, _ := primitive.ObjectIDFromHex(userIDStr)
	_, err = s.repo.Update(ctx, report.ID, userObjID, bson.M{"$set": report})
	return report, err
}

func (s *reportService) RemoveIncome(ctx context.Context, reportID, userIDStr, incomeID string) (*models.Report, error) {
	report, err := s.GetReportByID(ctx, reportID, userIDStr)
	if err != nil {
		return nil, err
	}

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
		return nil, errors.New("income not found")
	}

	report.Ingresos = updatedIncomes
	recalcReportTotals(report)

	userObjID, _ := primitive.ObjectIDFromHex(userIDStr)
	_, err = s.repo.Update(ctx, report.ID, userObjID, bson.M{"$set": report})
	return report, err
}

func (s *reportService) RemoveExpense(ctx context.Context, reportID, userIDStr, expenseID string) (*models.Report, error) {
	report, err := s.GetReportByID(ctx, reportID, userIDStr)
	if err != nil {
		return nil, err
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
		return nil, errors.New("expense not found")
	}

	report.Gastos = updatedExpenses
	recalcReportTotals(report)

	userObjID, _ := primitive.ObjectIDFromHex(userIDStr)
	_, err = s.repo.Update(ctx, report.ID, userObjID, bson.M{"$set": report})
	return report, err
}

func (s *reportService) GetAnnualReport(ctx context.Context, userIDStr string, year int) (bson.M, error) {
	userObjID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	// Copiamos el pipeline original
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

	results, err := s.repo.AggregateAnnual(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}

	result := results[0]
	// Aplicar redondeo a los resultados
	fields := []string{"total_ingreso_bruto", "total_ingreso_neto", "total_diezmos", "total_ofrendas", "total_iglesia", "total_gastos", "liquidacion_final"}
	for _, f := range fields {
		if val, ok := result[f].(float64); ok {
			result[f] = roundToTwoDecimals(val)
		}
	}
	return result, nil
}
