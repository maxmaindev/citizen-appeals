package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrServiceEmbeddingNotFound = errors.New("service embedding not found")
)

type ServiceEmbeddingRepository struct {
	collection *mongo.Collection
}

func NewServiceEmbeddingRepository(collection *mongo.Collection) *ServiceEmbeddingRepository {
	return &ServiceEmbeddingRepository{
		collection: collection,
	}
}

type ServiceEmbedding struct {
	ServiceID   int64     `bson:"service_id"`
	ServiceName string    `bson:"service_name"`
	Description string    `bson:"description"`
	Examples    []string  `bson:"examples"`
	Version     int       `bson:"version"`
	UpdatedAt   time.Time `bson:"updated_at"`
	CreatedAt   time.Time `bson:"created_at"`
}

// GetByServiceID retrieves service embedding by service ID
func (r *ServiceEmbeddingRepository) GetByServiceID(ctx context.Context, serviceID int64) (*ServiceEmbedding, error) {
	filter := bson.M{"service_id": serviceID}
	
	var embedding ServiceEmbedding
	err := r.collection.FindOne(ctx, filter).Decode(&embedding)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrServiceEmbeddingNotFound
		}
		return nil, fmt.Errorf("failed to get service embedding: %w", err)
	}
	
	return &embedding, nil
}

// GetByServiceName retrieves service embedding by service name
func (r *ServiceEmbeddingRepository) GetByServiceName(ctx context.Context, serviceName string) (*ServiceEmbedding, error) {
	filter := bson.M{"service_name": serviceName}
	
	var embedding ServiceEmbedding
	err := r.collection.FindOne(ctx, filter).Decode(&embedding)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrServiceEmbeddingNotFound
		}
		return nil, fmt.Errorf("failed to get service embedding: %w", err)
	}
	
	return &embedding, nil
}

// UpdateKeywords updates keywords (examples) for a service
func (r *ServiceEmbeddingRepository) UpdateKeywords(ctx context.Context, serviceID int64, serviceName string, keywords string) error {
	// Split keywords by semicolon and clean up
	examples := []string{}
	if keywords != "" {
		parts := strings.Split(keywords, ";")
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				examples = append(examples, trimmed)
			}
		}
	}

	filter := bson.M{"service_id": serviceID}
	
	// Check if document exists
	var existing ServiceEmbedding
	err := r.collection.FindOne(ctx, filter).Decode(&existing)
	
	if err == mongo.ErrNoDocuments {
		// Create new document
		embedding := ServiceEmbedding{
			ServiceID:   serviceID,
			ServiceName: serviceName,
			Description: "", // Will be updated from PostgreSQL if needed
			Examples:    examples,
			Version:     1,
			UpdatedAt:   time.Now(),
			CreatedAt:   time.Now(),
		}
		
		_, err = r.collection.InsertOne(ctx, embedding)
		if err != nil {
			return fmt.Errorf("failed to create service embedding: %w", err)
		}
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to check existing embedding: %w", err)
	}
	
	// Update existing document
	update := bson.M{
		"$set": bson.M{
			"service_name": serviceName,
			"examples":     examples,
			"version":      existing.Version + 1,
			"updated_at":   time.Now(),
		},
	}
	
	_, err = r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update service embedding: %w", err)
	}
	
	return nil
}

// GetAllForClassification returns all service embeddings for ML classification
func (r *ServiceEmbeddingRepository) GetAllForClassification(ctx context.Context) (map[string]string, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to find service embeddings: %w", err)
	}
	defer cursor.Close(ctx)
	
	services := make(map[string]string)
	
	for cursor.Next(ctx) {
		var embedding ServiceEmbedding
		if err := cursor.Decode(&embedding); err != nil {
			continue
		}
		
		// Combine description and examples for classification
		combinedDescription := embedding.Description
		if len(embedding.Examples) > 0 {
			examplesText := strings.Join(embedding.Examples, "; ")
			if combinedDescription != "" {
				combinedDescription += "; " + examplesText
			} else {
				combinedDescription = examplesText
			}
		}
		
		services[embedding.ServiceName] = combinedDescription
	}
	
	return services, nil
}

// GetKeywords returns keywords (examples) for a service as a string (semicolon-separated)
func (r *ServiceEmbeddingRepository) GetKeywords(ctx context.Context, serviceID int64) (string, error) {
	embedding, err := r.GetByServiceID(ctx, serviceID)
	if err != nil {
		if err == ErrServiceEmbeddingNotFound {
			return "", nil // Return empty string if not found
		}
		return "", err
	}
	
	return strings.Join(embedding.Examples, "; "), nil
}

// CreateIndexes creates necessary indexes
func (r *ServiceEmbeddingRepository) CreateIndexes(ctx context.Context) error {
	// Index on service_id (unique)
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "service_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := r.collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("failed to create service_id index: %w", err)
	}
	
	// Index on service_name
	indexModel = mongo.IndexModel{
		Keys: bson.D{{Key: "service_name", Value: 1}},
	}
	_, err = r.collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("failed to create service_name index: %w", err)
	}
	
	return nil
}

