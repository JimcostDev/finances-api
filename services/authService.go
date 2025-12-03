package services

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/JimcostDev/finances-api/models"
	"github.com/JimcostDev/finances-api/repositories"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	RegisterUser(ctx context.Context, req RegisterRequest) (*models.User, error)
	LoginUser(ctx context.Context, req LoginRequest) (string, error)
}

// Estructuras de Request (Movidas aquí para ser accesibles)
type RegisterRequest struct {
	Email           string `json:"email"`
	Username        string `json:"username"`
	Fullname        string `json:"fullname"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authService struct {
	repo repositories.UserRepository
}

func NewAuthService(repo repositories.UserRepository) AuthService {
	return &authService{repo: repo}
}

func (s *authService) RegisterUser(ctx context.Context, req RegisterRequest) (*models.User, error) {
	// 1. Validar contraseñas
	if req.Password != req.ConfirmPassword {
		return nil, errors.New("las contraseñas no coinciden")
	}

	// 2. Validar existencia
	exists, err := s.repo.ExistsByEmailOrUsername(ctx, req.Email, req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("el email o el username ya existen")
	}

	// 3. Hashear password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("error al encriptar la contraseña")
	}

	// 4. Crear modelo
	user := models.User{
		Email:     req.Email,
		Username:  req.Username,
		Fullname:  req.Fullname,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 5. Guardar en DB
	_, err = s.repo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	// Limpiar password antes de devolver
	user.Password = ""
	return &user, nil
}

func (s *authService) LoginUser(ctx context.Context, req LoginRequest) (string, error) {
	// 1. Buscar usuario
	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return "", errors.New("credenciales inválidas")
	}

	// 2. Comparar hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", errors.New("credenciales inválidas")
	}

	// 3. Generar JWT
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = user.ID
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(72 * time.Hour).Unix()

	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		return "", errors.New("JWT_SECRET_KEY no está definida")
	}

	t, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", errors.New("no se pudo generar el token")
	}

	return t, nil
}
