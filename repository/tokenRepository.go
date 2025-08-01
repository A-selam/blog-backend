package repository

import (
	"blog-backend/domain"
	"context"

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
	res, err := tr.database.Collection(tr.collection).DeleteOne(ctx, map[string]interface{}{"token": token})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0{
		return domain.ErrTokenNotFound
	}
	return nil
}

func (tr *refreshTokenRepository) DeleteRefreshTokensForUser(ctx context.Context, userID string) error {

	res, err := tr.database.Collection(tr.collection).DeleteMany(ctx, map[string]interface{}{"userID": userID})
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

// Password Reset Tokens
func (tr *resetTokenRepository) CreatePasswordResetToken(ctx context.Context, token *domain.PasswordResetToken) (*domain.PasswordResetToken, error) {
	// TODO: implement this function
	return nil, nil
}

func (tr *resetTokenRepository) GetPasswordResetToken(ctx context.Context, token string) (*domain.PasswordResetToken, error) {
	// TODO: implement this function
	return nil, nil
}

func (tr *resetTokenRepository) MarkPasswordResetTokenUsed(ctx context.Context, token string) error {
	// TODO: implement this function
	return nil
}
