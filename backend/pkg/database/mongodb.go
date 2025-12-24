package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"citizen-appeals/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func NewMongoDB(cfg *config.Config) (*MongoDB, error) {
	if !cfg.MongoDB.Enabled {
		return nil, fmt.Errorf("MongoDB is disabled")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Prepare URI
	uri := cfg.MongoDB.URI
	if !strings.Contains(uri, "authSource") {
		// Add authSource to URI if not present
		separator := "&"
		if !strings.Contains(uri, "?") {
			separator = "?"
			// Check if URI needs / before ?
			lastColon := strings.LastIndex(uri, ":")
			hasSlashAfterPort := lastColon != -1 && strings.Contains(uri[lastColon:], "/")
			if !hasSlashAfterPort {
				uri += "/"
			}
		}
		uri += separator + "authSource=" + cfg.MongoDB.AuthSource
	}

	// Create MongoDB client
	clientOptions := options.Client().ApplyURI(uri)
	clientOptions.SetServerSelectionTimeout(5 * time.Second)
	clientOptions.SetConnectTimeout(10 * time.Second)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test connection
	if err := client.Ping(ctx, nil); err != nil {
		client.Disconnect(ctx)
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(cfg.MongoDB.Database)

	return &MongoDB{
		Client:   client,
		Database: database,
	}, nil
}

func (m *MongoDB) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return m.Client.Disconnect(ctx)
}

func (m *MongoDB) Ping(ctx context.Context) error {
	return m.Client.Ping(ctx, nil)
}

