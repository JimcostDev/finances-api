package handlers

import (
	"context"
	"time"

	"github.com/JimcostDev/finances-api/config"
	"github.com/JimcostDev/finances-api/models"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

// DeleteUser elimina un usuario y todos sus reportes
func DeleteUser(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Usuario no autenticado"})
	}

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID inválido"})
	}

	// Obtener cliente MongoDB desde la configuración
	client := config.GetClient()

	// Crear sesión
	session, err := client.StartSession()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error al iniciar sesión"})
	}
	defer session.EndSession(context.Background())

	// Transacción
	err = mongo.WithSession(context.Background(), session, func(sessionContext mongo.SessionContext) error {
		if err := session.StartTransaction(); err != nil {
			return err
		}

		// 1. Eliminar usuario
		usersCollection := config.DB.Collection("users")
		if _, err := usersCollection.DeleteOne(sessionContext, bson.M{"_id": objID}); err != nil {
			session.AbortTransaction(sessionContext)
			return err
		}

		// 2. Eliminar reportes
		reportsCollection := config.DB.Collection("reports")
		if _, err := reportsCollection.DeleteMany(sessionContext, bson.M{"user_id": objID}); err != nil {
			session.AbortTransaction(sessionContext)
			return err
		}

		return session.CommitTransaction(sessionContext)
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error en transacción: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Usuario y reportes asociados eliminados exitosamente",
	})
}
