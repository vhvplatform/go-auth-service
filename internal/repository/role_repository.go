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

// RoleRepository handles role data access
type RoleRepository struct {
	collection *mongo.Collection
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(db *mongo.Database) *RoleRepository {
	collection := db.Collection("roles")

	// Create indexes
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "name", Value: 1},
				{Key: "tenantId", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	}

	_, _ = collection.Indexes().CreateMany(ctx, indexes)

	return &RoleRepository{collection: collection}
}

// FindByNames finds roles by their names
func (r *RoleRepository) FindByNames(ctx context.Context, names []string, tenantID string) ([]*domain.Role, error) {
	filter := bson.M{
		"name": bson.M{"$in": names},
		"$or": []bson.M{
			{"tenantId": tenantID},
			{"tenantId": bson.M{"$exists": false}},
		},
	}

	// Use hint to leverage the compound index for better performance
	opts := options.Find().SetHint(bson.D{{Key: "name", Value: 1}, {Key: "tenantId", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find roles: %w", err)
	}
	defer cursor.Close(ctx)

	var roles []*domain.Role
	if err := cursor.All(ctx, &roles); err != nil {
		return nil, fmt.Errorf("failed to decode roles: %w", err)
	}

	return roles, nil
}

// GetPermissionsForRoles gets all permissions for a set of roles
func (r *RoleRepository) GetPermissionsForRoles(ctx context.Context, roles []string, tenantID string) ([]string, error) {
	foundRoles, err := r.FindByNames(ctx, roles, tenantID)
	if err != nil {
		return nil, err
	}

	permissionsMap := make(map[string]bool)
	for _, role := range foundRoles {
		for _, permission := range role.Permissions {
			permissionsMap[permission] = true
		}
	}

	permissions := make([]string, 0, len(permissionsMap))
	for permission := range permissionsMap {
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

// Create creates a new role
func (r *RoleRepository) Create(ctx context.Context, role *domain.Role) error {
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, role)
	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		role.ID = oid
	}
	return nil
}

// FindByNameAndTenant finds a role by name and tenant
func (r *RoleRepository) FindByNameAndTenant(ctx context.Context, name, tenantID string) (*domain.Role, error) {
	var role domain.Role
	err := r.collection.FindOne(ctx, bson.M{
		"name":     name,
		"tenantId": tenantID,
	}).Decode(&role)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find role: %w", err)
	}
	return &role, nil
}
