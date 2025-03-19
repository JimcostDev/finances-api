package handlers

import (
	"context"
	"os"
	"time"

	"github.com/JimcostDev/finances-api/config"
	"github.com/JimcostDev/finances-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// Register crea un nuevo usuario en la base de datos, validando que el email y username sean únicos.
func Register(c *fiber.Ctx) error {
	// Define un struct de solicitud para parsear el JSON
	type RegisterRequest struct {
		Email           string `json:"email"`
		Username        string `json:"username"`
		Fullname        string `json:"fullname"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
	}

	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Error al parsear JSON"})
	}

	// Validar que las contraseñas coincidan
	if req.Password != req.ConfirmPassword {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Las contraseñas no coinciden"})
	}

	// Crea un usuario sin incluir confirm_password
	user := models.User{
		Email:    req.Email,
		Username: req.Username,
		Fullname: req.Fullname,
		Password: req.Password, // Se hasheará a continuación
	}

	collection := config.DB.Collection("users")
	// Verificar si ya existe un usuario con el mismo email o username
	filter := bson.M{
		"$or": []bson.M{
			{"email": user.Email},
			{"username": user.Username},
		},
	}
	var existingUser models.User
	err := collection.FindOne(context.Background(), filter).Decode(&existingUser)
	if err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El email o el username ya existen"})
	} else if err != mongo.ErrNoDocuments {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error al validar usuario existente"})
	}

	// Hashear la contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error al encriptar la contraseña"})
	}
	user.Password = string(hashedPassword)
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err = collection.InsertOne(context.Background(), user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error al insertar el usuario"})
	}
	// No se devuelve la contraseña
	user.Password = ""
	return c.Status(fiber.StatusCreated).JSON(user)
}

// Login permite a un usuario iniciar sesión en la aplicación
func Login(c *fiber.Ctx) error {
	// Estructura para parsear la petición de login
	type LoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Error al parsear JSON"})
	}

	// Buscar el usuario por email en la colección "users"
	collection := config.DB.Collection("users")
	var user models.User
	err := collection.FindOne(context.Background(), bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Credenciales inválidas"})
	}

	// Comparar la contraseña hasheada con la proporcionada
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Credenciales inválidas"})
	}

	// Generar el token JWT
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = user.ID
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(72 * time.Hour).Unix() // Token con expiración de 72 horas

	// Obtener la clave secreta desde las variables de entorno
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "JWT_SECRET_KEY no está definida"})
	}
	// Firma el token con la clave secreta
	t, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "No se pudo generar el token"})
	}

	// Retornar el token en la respuesta
	return c.JSON(fiber.Map{"token": t})
}
