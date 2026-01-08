package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/vhvplatform/go-auth-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TenantRepository handles tenant data access
type TenantRepository struct {
	collection *mongo.Collection
}

// NewTenantRepository creates a new tenant repository
func NewTenantRepository(db *mongo.Database) *TenantRepository {
	collection := db.Collection("tenants")

	// Create indexes
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}

	_, _ = collection.Indexes().CreateMany(ctx, indexes)

	return &TenantRepository{collection: collection}
}

// Create creates a new tenant
func (r *TenantRepository) Create(ctx context.Context, tenant *domain.Tenant) error {
	tenant.CreatedAt = time.Now()
	tenant.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, tenant)
	if err != nil {
		return fmt.Errorf("failed to create tenant: %w", err)
	}

	return nil
}

// FindByID finds a tenant by ID
func (r *TenantRepository) FindByID(ctx context.Context, id string) (*domain.Tenant, error) {
	var tenant domain.Tenant
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&tenant)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find tenant by ID: %w", err)
	}
	return &tenant, nil
}

// Update updates a tenant
func (r *TenantRepository) Update(ctx context.Context, tenant *domain.Tenant) error {
	tenant.UpdatedAt = time.Now()

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": tenant.ID},
		bson.M{"$set": tenant},
	)
	if err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}
	return nil
}

// ListActive lists active tenants
func (r *TenantRepository) ListActive(ctx context.Context) ([]*domain.Tenant, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"isActive": true})
	if err != nil {
		return nil, fmt.Errorf("failed to list active tenants: %w", err)
	}
	defer cursor.Close(ctx)

	var tenants []*domain.Tenant
	if err := cursor.All(ctx, &tenants); err != nil {
		return nil, fmt.Errorf("failed to decode tenants: %w", err)
	}
	return tenants, nil
}
