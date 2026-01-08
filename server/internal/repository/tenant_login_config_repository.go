package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/vhvplatform/go-auth-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TenantLoginConfigRepository handles tenant login configuration data access
type TenantLoginConfigRepository struct {
	collection *mongo.Collection
}

// NewTenantLoginConfigRepository creates a new tenant login config repository
func NewTenantLoginConfigRepository(db *mongo.Database) *TenantLoginConfigRepository {
	collection := db.Collection("tenant_login_configs")

	// Create indexes
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "tenantId", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}

	_, _ = collection.Indexes().CreateMany(ctx, indexes)

	return &TenantLoginConfigRepository{collection: collection}
}

// Create creates a new tenant login configuration
func (r *TenantLoginConfigRepository) Create(ctx context.Context, config *domain.TenantLoginConfig) error {
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()

	// Set defaults if not provided
	if len(config.AllowedIdentifiers) == 0 {
		config.AllowedIdentifiers = []string{"email", "username"}
	}
	if config.PasswordMinLength == 0 {
		config.PasswordMinLength = 8
	}
	if config.SessionTimeout == 0 {
		config.SessionTimeout = 1440 // 24 hours
	}
	if config.MaxLoginAttempts == 0 {
		config.MaxLoginAttempts = 5
	}
	if config.LockoutDuration == 0 {
		config.LockoutDuration = 30 // 30 minutes
	}

	result, err := r.collection.InsertOne(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to create tenant login config: %w", err)
	}

	config.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindByTenant finds login configuration for a tenant
func (r *TenantLoginConfigRepository) FindByTenant(ctx context.Context, tenantID string) (*domain.TenantLoginConfig, error) {
	var config domain.TenantLoginConfig
	filter := bson.M{"tenantId": tenantID}

	err := r.collection.FindOne(ctx, filter).Decode(&config)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Return default config if not found
			return r.GetDefaultConfig(tenantID), nil
		}
		return nil, fmt.Errorf("failed to find tenant login config: %w", err)
	}
	return &config, nil
}

// Update updates tenant login configuration
func (r *TenantLoginConfigRepository) Update(ctx context.Context, config *domain.TenantLoginConfig) error {
	config.UpdatedAt = time.Now()

	filter := bson.M{"tenantId": config.TenantID}
	update := bson.M{"$set": config}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update tenant login config: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("tenant login config not found")
	}

	return nil
}

// Upsert creates or updates tenant login configuration
func (r *TenantLoginConfigRepository) Upsert(ctx context.Context, config *domain.TenantLoginConfig) error {
	config.UpdatedAt = time.Now()
	if config.CreatedAt.IsZero() {
		config.CreatedAt = time.Now()
	}

	filter := bson.M{"tenantId": config.TenantID}
	update := bson.M{
		"$set": config,
		"$setOnInsert": bson.M{
			"createdAt": config.CreatedAt,
		},
	}

	opts := options.Update().SetUpsert(true)
	result, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to upsert tenant login config: %w", err)
	}

	if result.UpsertedID != nil {
		config.ID = result.UpsertedID.(primitive.ObjectID)
	}

	return nil
}

// Delete deletes tenant login configuration
func (r *TenantLoginConfigRepository) Delete(ctx context.Context, tenantID string) error {
	filter := bson.M{"tenantId": tenantID}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete tenant login config: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("tenant login config not found")
	}

	return nil
}

// GetDefaultConfig returns a default login configuration
func (r *TenantLoginConfigRepository) GetDefaultConfig(tenantID string) *domain.TenantLoginConfig {
	return &domain.TenantLoginConfig{
		TenantID:             tenantID,
		AllowedIdentifiers:   []string{"email", "username"},
		Require2FA:           false,
		AllowRegistration:    true,
		PasswordMinLength:    8,
		PasswordRequireUpper: true,
		PasswordRequireLower: true,
		PasswordRequireDigit: true,
		PasswordRequireSpec:  false,
		SessionTimeout:       1440, // 24 hours
		MaxLoginAttempts:     5,
		LockoutDuration:      30, // 30 minutes
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
}

// IsIdentifierAllowed checks if an identifier type is allowed for login in this tenant
func (r *TenantLoginConfigRepository) IsIdentifierAllowed(ctx context.Context, tenantID string, identifierType string) (bool, error) {
	config, err := r.FindByTenant(ctx, tenantID)
	if err != nil {
		return false, err
	}

	for _, allowed := range config.AllowedIdentifiers {
		if allowed == identifierType {
			return true, nil
		}
	}

	return false, nil
}
