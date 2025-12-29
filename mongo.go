package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Variable global para la conexión (Pool de conexiones)
var collection *mongo.Collection

// Función init o main para conectar
func initDB() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal("Cant connect to MongoDB", err)
	}

	fmt.Println("Conectado a MongoDB local!")

	collection = client.Database("shortenerDB").Collection("urls")

}
