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

// RefreshTokenRepository handles refresh token data access
type RefreshTokenRepository struct {
	collection *mongo.Collection
}

// NewRefreshTokenRepository creates a new refresh token repository
func NewRefreshTokenRepository(db *mongo.Database) *RefreshTokenRepository {
	collection := db.Collection("refresh_tokens")

	// Create indexes
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "userId", Value: 1}},
		},
		{
			Keys:    bson.D{{Key: "token", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "expiresAt", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0),
		},
	}

	_, _ = collection.Indexes().CreateMany(ctx, indexes)

	return &RefreshTokenRepository{collection: collection}
}

// Create creates a new refresh token
func (r *RefreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	token.CreatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	token.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindByToken finds a refresh token by token string
func (r *RefreshTokenRepository) FindByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	var refreshToken domain.RefreshToken
	// Optimize query with projection
	opts := options.FindOne().SetProjection(bson.M{
		"_id":       1,
		"userId":    1,
		"token":     1,
		"expiresAt": 1,
		"createdAt": 1,
		"revokedAt": 1,
	})
	err := r.collection.FindOne(ctx, bson.M{
		"token":     token,
		"revokedAt": nil,
		"expiresAt": bson.M{"$gt": time.Now()},
	}, opts).Decode(&refreshToken)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find refresh token: %w", err)
	}
	return &refreshToken, nil
}

// Revoke revokes a refresh token
func (r *RefreshTokenRepository) Revoke(ctx context.Context, token string) error {
	now := time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"token": token},
		bson.M{"$set": bson.M{"revokedAt": now}},
	)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}
	return nil
}

// RevokeAllForUser revokes all refresh tokens for a user
func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	now := time.Now()
	_, err := r.collection.UpdateMany(
		ctx,
		bson.M{"userId": userID, "revokedAt": nil},
		bson.M{"$set": bson.M{"revokedAt": now}},
	)
	if err != nil {
		return fmt.Errorf("failed to revoke user tokens: %w", err)
	}
	return nil
}

// DeleteExpiredTokens removes all expired and revoked tokens (for manual cleanup)
func (r *RefreshTokenRepository) DeleteExpiredTokens(ctx context.Context) (int64, error) {
	result, err := r.collection.DeleteMany(
		ctx,
		bson.M{
			"$or": []bson.M{
				{"expiresAt": bson.M{"$lt": time.Now()}},
				{"revokedAt": bson.M{"$ne": nil}},
			},
		},
	)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired tokens: %w", err)
	}
	return result.DeletedCount, nil
}

// CountActiveTokensForUser returns the number of active tokens for a user
func (r *RefreshTokenRepository) CountActiveTokensForUser(ctx context.Context, userID string) (int64, error) {
	count, err := r.collection.CountDocuments(
		ctx,
		bson.M{
			"userId":    userID,
			"revokedAt": nil,
			"expiresAt": bson.M{"$gt": time.Now()},
		},
	)
	if err != nil {
		return 0, fmt.Errorf("failed to count active tokens: %w", err)
	}
	return count, nil
}
