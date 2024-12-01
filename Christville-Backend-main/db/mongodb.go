package db

import (
	"context"
	"crypto/tls"
	"log"
	"sync"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var clientInstance *mongo.Client
var clientInstanceError error
var mongoOnce sync.Once

// ConnectMongoDB initializes a connection to the MongoDB server with TLS
func ConnectMongoDB() (*mongo.Client, error) {

	err := godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }

    // Get the MongoDB URI from environment variables
    mongoURI := os.Getenv("MONGODB_URI")
    if mongoURI == "" {
        log.Fatal("MONGODB_URI is not set in the environment")
    }

	mongoOnce.Do(func() {
		// Set client options with TLS configuration
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}

		clientOptions := options.Client().
			ApplyURI(mongoURI).
			SetTLSConfig(tlsConfig)

		client, err := mongo.Connect(context.TODO(), clientOptions)
		if err != nil {
			clientInstanceError = err
			return
		}

		err = client.Ping(context.TODO(), nil)
		if err != nil {
			clientInstanceError = err
			return
		}

		clientInstance = client
		log.Println("Connected to MongoDB successfully")
	})

	return clientInstance, clientInstanceError
}

// GetClient returns the MongoDB client instance (singleton)
func GetClient() *mongo.Client {
	return clientInstance
}
