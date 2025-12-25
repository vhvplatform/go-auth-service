package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/vhvcorp/go-auth-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
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
				{Key: "tenant_id", Value: 1},
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
			{"tenant_id": tenantID},
			{"tenant_id": bson.M{"$exists": false}},
		},
	}
	
	cursor, err := r.collection.Find(ctx, filter)
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
