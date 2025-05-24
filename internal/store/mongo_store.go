package store

import (
	"context"
	"errors"
	"fmt"
	"log"

	"shawty/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// UrlStoreInterface defines the operations for URL persistence.
type UrlStoreInterface interface {
	Save(ctx context.Context, urlEntry domain.URL) error
	GetByShortID(ctx context.Context, shortID string) (domain.URL, error)
	EnsureIndexes(ctx context.Context) error
}

// MongoUrlStore implements UrlStoreInterface using MongoDB.
type MongoUrlStore struct {
	collection *mongo.Collection
}

// NewMongoUrlStore creates a new MongoUrlStore.
func NewMongoUrlStore(dbClient *mongo.Client, dbName string, collectionName string) *MongoUrlStore {
	collection := dbClient.Database(dbName).Collection(collectionName)
	return &MongoUrlStore{collection: collection}
}

// EnsureIndexes creates necessary indexes for the urls collection.
// For example, a unique index on the short_url field (or _id if it\'s the same).
func (s *MongoUrlStore) EnsureIndexes(ctx context.Context) error {
	// The _id field is automatically indexed by MongoDB and is always unique.
	// Explicitly trying to create it again with SetUnique(true) causes the reported error.
	// If other fields needed indexing (e.g., a non-unique index on 'original_url' for searching,
	// or a unique index on a 'short_url' field if it were different from _id),
	// those definitions would go here.

	// Example of creating an index on another field, if needed:
	/*
		otherFieldIndexModel := mongo.IndexModel{
			Keys:    bson.D{{Key: "some_other_field", Value: 1}},
			Options: options.Index().SetUnique(false), // Or true if it needs to be unique
		}
		_, err := s.collection.Indexes().CreateOne(ctx, otherFieldIndexModel)
		if err != nil {
			// It's good practice to check if the error is because the index already exists.
			// For example, by checking the error code or message.
			// if mongo.IsDuplicateKeyError(err) || strings.Contains(err.Error(), "index already exists") {
			// log.Println("Index on 'some_other_field' already exists.")
			// } else {
			return fmt.Errorf("failed to create index on some_other_field: %w", err)
			// }
		} else {
			log.Println("Successfully ensured index on some_other_field.")
		}
	*/

	log.Println("MongoDB automatically handles the unique index on _id. EnsureIndexes will ensure other custom indexes if defined above.")
	return nil
}

// Save inserts a new URL entry into the database.
func (s *MongoUrlStore) Save(ctx context.Context, urlEntry domain.URL) error {
	_, err := s.collection.InsertOne(ctx, urlEntry)
	if err != nil {
		// Check for duplicate key error (code 11000)
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("short URL '%s' already exists: %w", urlEntry.ID, err)
		}
		return fmt.Errorf("failed to insert URL into MongoDB: %w", err)
	}
	return nil
}

// GetByShortID retrieves a URL entry by its short ID (_id field).
func (s *MongoUrlStore) GetByShortID(ctx context.Context, shortID string) (domain.URL, error) {
	var url domain.URL
	filter := bson.M{"_id": shortID}
	err := s.collection.FindOne(ctx, filter).Decode(&url)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.URL{}, fmt.Errorf("URL with ID '%s' not found: %w", shortID, err)
		}
		return domain.URL{}, fmt.Errorf("error retrieving URL from MongoDB: %w", err)
	}
	return url, nil
}
