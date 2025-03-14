package config

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database

func ConnectDB() {
	// Leer la variable de entorno MONGODB_URI
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		log.Fatal("MONGO_URI no está definida")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Usar mongo.Connect para conectarse a la base de datos
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	// Verifica la conexión
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Selecciona la base de datos que deseas utilizar
	DB = client.Database("finances")
	log.Println("Conectado a MongoDB!")
}
