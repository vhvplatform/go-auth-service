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

// UserTenantRepository handles user-tenant relationship data access
type UserTenantRepository struct {
	collection *mongo.Collection
}

// NewUserTenantRepository creates a new user-tenant repository
func NewUserTenantRepository(db *mongo.Database) *UserTenantRepository {
	collection := db.Collection("user_tenants")

	// Create indexes
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
				{Key: "tenantId", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "userId", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "tenantId", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "isActive", Value: 1}},
		},
	}

	_, _ = collection.Indexes().CreateMany(ctx, indexes)

	return &UserTenantRepository{collection: collection}
}

// Create creates a new user-tenant relationship
func (r *UserTenantRepository) Create(ctx context.Context, userTenant *domain.UserTenant) error {
	userTenant.CreatedAt = time.Now()
	userTenant.UpdatedAt = time.Now()
	userTenant.JoinedAt = time.Now()
	userTenant.IsActive = true

	result, err := r.collection.InsertOne(ctx, userTenant)
	if err != nil {
		return fmt.Errorf("failed to create user-tenant relationship: %w", err)
	}

	userTenant.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindByUserAndTenant finds a user-tenant relationship
func (r *UserTenantRepository) FindByUserAndTenant(ctx context.Context, userID, tenantID string) (*domain.UserTenant, error) {
	var userTenant domain.UserTenant
	filter := bson.M{
		"userId":   userID,
		"tenantId": tenantID,
	}

	err := r.collection.FindOne(ctx, filter).Decode(&userTenant)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user-tenant relationship: %w", err)
	}
	return &userTenant, nil
}

// FindByUser finds all tenant relationships for a user
func (r *UserTenantRepository) FindByUser(ctx context.Context, userID string) ([]*domain.UserTenant, error) {
	filter := bson.M{"userId": userID, "isActive": true}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find user tenants: %w", err)
	}
	defer cursor.Close(ctx)

	var userTenants []*domain.UserTenant
	if err := cursor.All(ctx, &userTenants); err != nil {
		return nil, fmt.Errorf("failed to decode user tenants: %w", err)
	}
	return userTenants, nil
}

// FindByTenant finds all users in a tenant
func (r *UserTenantRepository) FindByTenant(ctx context.Context, tenantID string, limit, skip int64) ([]*domain.UserTenant, error) {
	filter := bson.M{"tenantId": tenantID, "isActive": true}

	opts := options.Find().SetLimit(limit).SetSkip(skip).SetSort(bson.D{{Key: "joinedAt", Value: -1}})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find tenant users: %w", err)
	}
	defer cursor.Close(ctx)

	var userTenants []*domain.UserTenant
	if err := cursor.All(ctx, &userTenants); err != nil {
		return nil, fmt.Errorf("failed to decode tenant users: %w", err)
	}
	return userTenants, nil
}

// UpdateRoles updates the roles for a user-tenant relationship
func (r *UserTenantRepository) UpdateRoles(ctx context.Context, userID, tenantID string, roles []string) error {
	filter := bson.M{
		"userId":   userID,
		"tenantId": tenantID,
	}

	update := bson.M{
		"$set": bson.M{
			"roles":     roles,
			"updatedAt": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update user-tenant roles: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("user-tenant relationship not found")
	}

	return nil
}

// Deactivate deactivates a user-tenant relationship
func (r *UserTenantRepository) Deactivate(ctx context.Context, userID, tenantID string) error {
	filter := bson.M{
		"userId":   userID,
		"tenantId": tenantID,
	}

	update := bson.M{
		"$set": bson.M{
			"isActive":  false,
			"updatedAt": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to deactivate user-tenant relationship: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("user-tenant relationship not found")
	}

	return nil
}

// Activate activates a user-tenant relationship
func (r *UserTenantRepository) Activate(ctx context.Context, userID, tenantID string) error {
	filter := bson.M{
		"userId":   userID,
		"tenantId": tenantID,
	}

	update := bson.M{
		"$set": bson.M{
			"isActive":  true,
			"updatedAt": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to activate user-tenant relationship: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("user-tenant relationship not found")
	}

	return nil
}

// Delete permanently removes a user-tenant relationship
func (r *UserTenantRepository) Delete(ctx context.Context, userID, tenantID string) error {
	filter := bson.M{
		"userId":   userID,
		"tenantId": tenantID,
	}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete user-tenant relationship: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("user-tenant relationship not found")
	}

	return nil
}

// CountByTenant counts users in a tenant
func (r *UserTenantRepository) CountByTenant(ctx context.Context, tenantID string) (int64, error) {
	filter := bson.M{"tenantId": tenantID, "isActive": true}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count tenant users: %w", err)
	}
	return count, nil
}
