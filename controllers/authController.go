package controllers

import (
	"context"
	"os"
	"time"

	"github.com/JimcostDev/finances-api/config"
	"github.com/JimcostDev/finances-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

// Register crea un nuevo usuario en la base de datos
func Register(c *fiber.Ctx) error {
	user := new(models.User)
	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Error al parsear JSON"})
	}

	// Hashear la contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error al encriptar la contraseña"})
	}
	user.Password = string(hashedPassword)
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	collection := config.DB.Collection("users")
	_, err = collection.InsertOne(context.Background(), user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error al insertar el usuario"})
	}
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
