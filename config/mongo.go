package config

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Client *mongo.Client
	DB     *mongo.Database
)

func ConnectDB() {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		log.Fatal("MONGO_URI no está definida")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Configurar opciones del cliente para transacciones
	clientOpts := options.Client().ApplyURI(uri).SetRetryWrites(true)

	var err error
	Client, err = mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatal(err)
	}

	// Verificar conexión
	err = Client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Configurar la base de datos
	DB = Client.Database("finances")
	log.Println("Conectado a MongoDB!")
}

// GetClient devuelve la instancia del cliente MongoDB
func GetClient() *mongo.Client {
	return Client
}
