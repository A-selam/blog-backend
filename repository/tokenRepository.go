package repository

import (
	"blog-backend/domain"
	"context"
	"time"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type refreshTokenRepository struct {
	database   *mongo.Database
	collection string
}

func NewRefreshTokenRepositoryFromDB(db *mongo.Database) domain.IRefreshTokenRepository {
	return &refreshTokenRepository{
		database:   db,
		collection: "refreshTokens",
	}
}

// Refresh Tokens
func (tr *refreshTokenRepository) CreateRefreshToken(ctx context.Context, token *domain.RefreshToken) (*domain.RefreshToken, error) {
	// TODO: implement this function
	return nil, nil
}

func (tr *refreshTokenRepository) GetRefreshToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	// TODO: implement this function
	return nil, nil
}

func (tr *refreshTokenRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	res, err := tr.database.Collection(tr.collection).DeleteOne(ctx,bson.M{"token": token})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0{
		return domain.ErrTokenNotFound
	}
	return nil
}

func (tr *refreshTokenRepository) DeleteRefreshTokensForUser(ctx context.Context, userID string) error {

	res, err := tr.database.Collection(tr.collection).DeleteMany(ctx, bson.M{"user_id": userID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0{
		return domain.ErrTokenNotFound
	}
	return nil
}

type resetTokenRepository struct {
	database   *mongo.Database
	collection string
}

func NewResetTokenRepository(db *mongo.Database) domain.IResetTokenRepository {
	return &resetTokenRepository{
		database:   db,
		collection: "refreshTokens",
	}
}

func (tr *resetTokenRepository) CreatePasswordResetToken(ctx context.Context, token *domain.PasswordResetToken) (*domain.PasswordResetToken, error) {
    if token.CreatedAt.IsZero() {
        token.CreatedAt = time.Now()
    }
    if !token.Used {
        token.Used = false
    }

    _, err := tr.database.Collection(tr.collection).InsertOne(ctx, token)
    if err != nil {
        return nil, err
    }
    return token, nil
}

func (tr *resetTokenRepository) GetPasswordResetToken(ctx context.Context, token string) (*domain.PasswordResetToken, error) {
    var result domain.PasswordResetToken
    err := tr.database.Collection(tr.collection).FindOne(ctx, bson.M{"token": token}).Decode(&result)
    if err == mongo.ErrNoDocuments {
        return nil, domain.ErrTokenNotFound
    }
    if err != nil {
        return nil, err
    }
    
    if result.Used {
        return nil, domain.ErrTokenUsed
    }
    
    if !result.ExpiresAt.IsZero() && result.ExpiresAt.Before(time.Now()) {
        return nil, domain.ErrTokenExpired
    }
    
    return &result, nil
}

func (tr *resetTokenRepository) MarkPasswordResetTokenUsed(ctx context.Context, token string) error {
    update := bson.M{
        "$set": bson.M{
            "used":      true,
            "updatedAt": time.Now(),
        },
    }
    
    result, err := tr.database.Collection(tr.collection).UpdateOne(ctx, bson.M{"token": token}, update)
    if err != nil {
        return err
    }
    if result.MatchedCount == 0 {
        return domain.ErrTokenNotFound
    }
    return nil
}