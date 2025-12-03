package repositories

import (
	"context"

	"github.com/JimcostDev/finances-api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// UserRepository define todas las operaciones de base de datos para usuarios
type UserRepository interface {
	// Métodos usados por Auth
	Create(ctx context.Context, user models.User) (*mongo.InsertOneResult, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	ExistsByEmailOrUsername(ctx context.Context, email, username string) (bool, error)

	// Métodos usados por User Profile
	FindByID(ctx context.Context, oid primitive.ObjectID) (*models.User, error)
	FindByUsername(ctx context.Context, username string) (*models.User, error)
	Update(ctx context.Context, oid primitive.ObjectID, update interface{}) (*mongo.UpdateResult, error)
	Delete(ctx context.Context, oid primitive.ObjectID) (*mongo.DeleteResult, error)
}

type userRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) UserRepository {
	return &userRepository{
		collection: db.Collection("users"),
	}
}

// --- Implementación de Métodos ---

// Create inserta un nuevo usuario
func (r *userRepository) Create(ctx context.Context, user models.User) (*mongo.InsertOneResult, error) {
	return r.collection.InsertOne(ctx, user)
}

// FindByEmail busca un usuario por su email
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ExistsByEmailOrUsername verifica si ya existe un usuario con ese email o username
func (r *userRepository) ExistsByEmailOrUsername(ctx context.Context, email, username string) (bool, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"email": email},
			{"username": username},
		},
	}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// FindByID busca un usuario por su ObjectID
func (r *userRepository) FindByID(ctx context.Context, oid primitive.ObjectID) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByUsername busca un usuario por su nombre de usuario (usado para validar duplicados al editar)
func (r *userRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update actualiza un usuario por su ID
func (r *userRepository) Update(ctx context.Context, oid primitive.ObjectID, update interface{}) (*mongo.UpdateResult, error) {
	return r.collection.UpdateOne(ctx, bson.M{"_id": oid}, update)
}

// Delete elimina un usuario por su ID
func (r *userRepository) Delete(ctx context.Context, oid primitive.ObjectID) (*mongo.DeleteResult, error) {
	return r.collection.DeleteOne(ctx, bson.M{"_id": oid})
}
