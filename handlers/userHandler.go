package handlers

import (
	"context"
	"time"

	"github.com/JimcostDev/finances-api/config"
	"github.com/JimcostDev/finances-api/models"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// GetUser obtiene un usuario por su ID
func GetUserProfile(c *fiber.Ctx) error {
	// Obtener el userID desde el token (c.Locals)
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Usuario no autenticado"})
	}

	// Convertir el userID a ObjectID
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	// Buscar usuario en la base de datos
	collection := config.DB.Collection("users")
	var user models.User
	err = collection.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Usuario no encontrado"})
	}

	// Omitir la contraseña en la respuesta
	user.Password = ""

	return c.JSON(user)
}

// UpdateUser actualiza la información de un usuario
func UpdateUser(c *fiber.Ctx) error {
	type UpdateUserRequest struct {
		Email           string `json:"email,omitempty"`
		Username        string `json:"username,omitempty"`
		Fullname        string `json:"fullname,omitempty"`
		Password        string `json:"password,omitempty"`
		ConfirmPassword string `json:"confirm_password,omitempty"`
	}

	// Obtener el userID desde el token
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Usuario no autenticado"})
	}

	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Error al parsear JSON"})
	}

	// Si se envía contraseña, validar que confirm_password coincida
	if req.Password != "" && req.Password != req.ConfirmPassword {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Las contraseñas no coinciden"})
	}

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	collection := config.DB.Collection("users")
	updateData := bson.M{"$set": bson.M{"updated_at": time.Now()}}

	// Verificar unicidad de email
	if req.Email != "" {
		var existingUser models.User
		err := collection.FindOne(context.Background(), bson.M{"email": req.Email, "_id": bson.M{"$ne": objID}}).Decode(&existingUser)
		if err == nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "El email ya está en uso"})
		}
		updateData["$set"].(bson.M)["email"] = req.Email
	}

	// Verificar unicidad de username
	if req.Username != "" {
		var existingUser models.User
		err := collection.FindOne(context.Background(), bson.M{"username": req.Username, "_id": bson.M{"$ne": objID}}).Decode(&existingUser)
		if err == nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "El nombre de usuario ya está en uso"})
		}
		updateData["$set"].(bson.M)["username"] = req.Username
	}

	// Actualizar fullname
	if req.Fullname != "" {
		updateData["$set"].(bson.M)["fullname"] = req.Fullname
	}

	// Actualizar contraseña si se proporcionó
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error al encriptar contraseña"})
		}
		updateData["$set"].(bson.M)["password"] = string(hashedPassword)
	}

	// Aplicar actualización
	result, err := collection.UpdateOne(context.Background(), bson.M{"_id": objID}, updateData)
	if err != nil || result.MatchedCount == 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error al actualizar usuario o usuario no encontrado"})
	}

	return c.JSON(fiber.Map{"message": "Usuario actualizado correctamente"})
}

// DeleteUser elimina un usuario
func DeleteUser(c *fiber.Ctx) error {
	// Obtener el userID desde el token
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Usuario no autenticado"})
	}

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	collection := config.DB.Collection("users")
	result, err := collection.DeleteOne(context.Background(), bson.M{"_id": objID})
	if err != nil || result.DeletedCount == 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error al eliminar usuario o usuario no encontrado"})
	}

	return c.JSON(fiber.Map{"message": "Usuario eliminado correctamente"})
}
