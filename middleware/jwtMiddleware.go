package middleware

import (
	"fmt"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

// Protected es un middleware para verificar la autenticación con JWT.
func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Obtener el token del encabezado Authorization
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Token no proporcionado"})
		}

		// Verificar si el token tiene el prefijo "Bearer "
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Formato de token inválido"})
		}

		// Leer la clave secreta desde las variables de entorno
		secretKey := os.Getenv("JWT_SECRET_KEY")
		if secretKey == "" {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "JWT_SECRET_KEY no configurada"})
		}

		// Verificar el token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("método de firma inválido")
			}
			return []byte(secretKey), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "No autorizado"})
		}

		// Extraer claims del token
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Claims inválidas"})
		}

		// Obtener el ID del usuario y almacenarlo en c.Locals
		userID, ok := claims["id"].(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "ID de usuario no válido en el token"})
		}
		//fmt.Println("UserID extraído del token:", userID) // Para depuración

		// Guardar el ID del usuario en c.Locals para usarlo en los controladores
		c.Locals("userID", userID)

		// Continuar con la siguiente función en la cadena de middleware
		return c.Next()
	}
}
