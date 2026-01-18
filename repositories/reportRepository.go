package repositories

import (
	"context"

	"github.com/JimcostDev/finances-api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Interfaz para definir qué hace el repositorio
type ReportRepository interface {
	Create(ctx context.Context, report models.Report) (*mongo.InsertOneResult, error)
	Update(ctx context.Context, oid primitive.ObjectID, userID primitive.ObjectID, update interface{}) (*mongo.UpdateResult, error)
	FindAll(ctx context.Context, userID primitive.ObjectID) ([]models.Report, error)
	FindOne(ctx context.Context, oid primitive.ObjectID, userID primitive.ObjectID) (*models.Report, error)
	FindByMonth(ctx context.Context, userID primitive.ObjectID, month string, year int) ([]models.Report, error)
	Delete(ctx context.Context, oid primitive.ObjectID, userID primitive.ObjectID) (*mongo.DeleteResult, error)
	AggregateReports(ctx context.Context, pipeline mongo.Pipeline) ([]bson.M, error)
	DeleteAllByUserID(ctx context.Context, userID primitive.ObjectID) (*mongo.DeleteResult, error)
}

type reportRepository struct {
	collection *mongo.Collection
}

func NewReportRepository(db *mongo.Database) ReportRepository {
	return &reportRepository{
		collection: db.Collection("reports"),
	}
}

// Implementación de métodos
func (r *reportRepository) Create(ctx context.Context, report models.Report) (*mongo.InsertOneResult, error) {
	return r.collection.InsertOne(ctx, report)
}

func (r *reportRepository) Update(ctx context.Context, oid primitive.ObjectID, userID primitive.ObjectID, update interface{}) (*mongo.UpdateResult, error) {
	filter := bson.M{"_id": oid, "user_id": userID}
	return r.collection.UpdateOne(ctx, filter, update)
}

func (r *reportRepository) FindAll(ctx context.Context, userID primitive.ObjectID) ([]models.Report, error) {
	filter := bson.M{"user_id": userID}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reports []models.Report
	for cursor.Next(ctx) {
		var report models.Report
		if err := cursor.Decode(&report); err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}
	return reports, nil
}

func (r *reportRepository) FindOne(ctx context.Context, oid primitive.ObjectID, userID primitive.ObjectID) (*models.Report, error) {
	filter := bson.M{"_id": oid, "user_id": userID}
	var report models.Report
	err := r.collection.FindOne(ctx, filter).Decode(&report)
	return &report, err
}

func (r *reportRepository) FindByMonth(ctx context.Context, userID primitive.ObjectID, month string, year int) ([]models.Report, error) {
	filter := bson.M{"user_id": userID, "month": month, "year": year}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reports []models.Report
	for cursor.Next(ctx) {
		var report models.Report
		if err := cursor.Decode(&report); err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}
	return reports, nil
}

func (r *reportRepository) Delete(ctx context.Context, oid primitive.ObjectID, userID primitive.ObjectID) (*mongo.DeleteResult, error) {
	filter := bson.M{"_id": oid, "user_id": userID}
	return r.collection.DeleteOne(ctx, filter)
}

func (r *reportRepository) AggregateReports(ctx context.Context, pipeline mongo.Pipeline) ([]bson.M, error) {
	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *reportRepository) DeleteAllByUserID(ctx context.Context, userID primitive.ObjectID) (*mongo.DeleteResult, error) {
	// Borra TODOS los documentos en la colección 'reports' que coincidan con el user_id
	return r.collection.DeleteMany(ctx, bson.M{"user_id": userID})
}
