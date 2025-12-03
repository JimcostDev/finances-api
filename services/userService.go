package services

import (
	"context"
	"errors"
	"time"

	"github.com/JimcostDev/finances-api/models"
	"github.com/JimcostDev/finances-api/repositories"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	GetUserProfile(ctx context.Context, userIDStr string) (*models.User, error)
	UpdateUser(ctx context.Context, userIDStr string, req UpdateUserRequest) error
	DeleteUser(ctx context.Context, userIDStr string) error
}

type UpdateUserRequest struct {
	Email           string `json:"email,omitempty"`
	Username        string `json:"username,omitempty"`
	Fullname        string `json:"fullname,omitempty"`
	Password        string `json:"password,omitempty"`
	ConfirmPassword string `json:"confirm_password,omitempty"`
}

type userService struct {
	userRepo   repositories.UserRepository
	reportRepo repositories.ReportRepository
	client     *mongo.Client // Necesario para transacciones
}

func NewUserService(uRepo repositories.UserRepository, rRepo repositories.ReportRepository, client *mongo.Client) UserService {
	return &userService{
		userRepo:   uRepo,
		reportRepo: rRepo,
		client:     client,
	}
}

func (s *userService) GetUserProfile(ctx context.Context, userIDStr string) (*models.User, error) {
	oid, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, errors.New("ID inválido")
	}

	user, err := s.userRepo.FindByID(ctx, oid)
	if err != nil {
		return nil, errors.New("usuario no encontrado")
	}
	user.Password = "" // Ocultar password
	return user, nil
}

func (s *userService) UpdateUser(ctx context.Context, userIDStr string, req UpdateUserRequest) error {
	oid, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return errors.New("ID inválido")
	}

	if req.Password != "" && req.Password != req.ConfirmPassword {
		return errors.New("las contraseñas no coinciden")
	}

	updateData := bson.M{"updated_at": time.Now()}

	// Validar Email único
	if req.Email != "" {
		existing, err := s.userRepo.FindByEmail(ctx, req.Email)
		if err == nil && existing.ID != oid {
			return errors.New("el email ya está en uso")
		}
		updateData["email"] = req.Email
	}

	// Validar Username único
	if req.Username != "" {
		existing, err := s.userRepo.FindByUsername(ctx, req.Username)
		if err == nil && existing.ID != oid {
			return errors.New("el nombre de usuario ya está en uso")
		}
		updateData["username"] = req.Username
	}

	if req.Fullname != "" {
		updateData["fullname"] = req.Fullname
	}

	if req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return errors.New("error al encriptar contraseña")
		}
		updateData["password"] = string(hashed)
	}

	_, err = s.userRepo.Update(ctx, oid, bson.M{"$set": updateData})
	return err
}

func (s *userService) DeleteUser(ctx context.Context, userIDStr string) error {
	oid, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return errors.New("ID inválido")
	}

	// Iniciar Sesión para Transacción
	session, err := s.client.StartSession()
	if err != nil {
		return errors.New("error al iniciar sesión de DB")
	}
	defer session.EndSession(ctx)

	// Ejecutar transacción
	err = mongo.WithSession(ctx, session, func(sessionContext mongo.SessionContext) error {
		if err := session.StartTransaction(); err != nil {
			return err
		}

		// 1. Eliminar Usuario
		if _, err := s.userRepo.Delete(sessionContext, oid); err != nil {
			session.AbortTransaction(sessionContext)
			return err
		}

		// 2. Eliminar Reportes asociados
		if _, err := s.reportRepo.DeleteAllByUserID(sessionContext, oid); err != nil {
			session.AbortTransaction(sessionContext)
			return err
		}

		return session.CommitTransaction(sessionContext)
	})

	return err
}
