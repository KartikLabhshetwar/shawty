package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DBConfig holds database configuration.
type DBConfig struct {
	URI            string
	DBName         string
	CollectionName string
	ConnectTimeout time.Duration
	PingTimeout    time.Duration
}

// LoadConfig loads database configuration from environment variables.
// Defaults are provided for some values.
func LoadConfig() DBConfig {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Info: .env file not found or error loading: %s. Using environment variables directly.", err)
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("MONGO_URI environment variable is required")
	}
	dbName := os.Getenv("MONGO_DB_NAME")
	if dbName == "" {
		dbName = "shawtydb"
		log.Printf("MONGO_DB_NAME not set, using default: %s", dbName)
	}
	collectionName := os.Getenv("MONGO_COLLECTION_NAME")
	if collectionName == "" {
		collectionName = "urls" // Default collection name
		log.Printf("MONGO_COLLECTION_NAME not set, using default: %s", collectionName)
	}

	return DBConfig{
		URI:            mongoURI,
		DBName:         dbName,
		CollectionName: collectionName,
		ConnectTimeout: 10 * time.Second,
		PingTimeout:    5 * time.Second,
	}
}

// ConnectDB establishes a connection to MongoDB using the provided configuration.
func ConnectDB(cfg DBConfig) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(cfg.URI)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectTimeout)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	ctxPing, cancelPing := context.WithTimeout(context.Background(), cfg.PingTimeout)
	defer cancelPing()
	err = client.Ping(ctxPing, nil)
	if err != nil {
		// Disconnect if ping fails to clean up resources
		if disconnectErr := client.Disconnect(context.Background()); disconnectErr != nil {
			log.Printf("Failed to disconnect client after ping failure: %v", disconnectErr)
		}
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	fmt.Println("Successfully connected to MongoDB!")
	return client, nil
}
