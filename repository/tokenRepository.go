package repository

import (
	"blog-backend/domain"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
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
func (tr *refreshTokenRepository) CreateRefreshToken(ctx context.Context, userID, refToken string) (*domain.RefreshToken, error) {
	collection := tr.database.Collection(tr.collection)

	refreshToken, err := refreshToken(userID, refToken)
	if err != nil {
		return nil, err	
	}

	insertResult, err := collection.InsertOne(ctx, refreshToken)
	if err != nil {	
		return nil, err
	}

	refreshToken.ID = insertResult.InsertedID.(bson.ObjectID)
	
	return &domain.RefreshToken{
		Token:     refreshToken.Token,
		ExpiresAt: refreshToken.ExpiresAt,
	}, nil
}

func (tr *refreshTokenRepository) ReplaceRefreshToken(ctx context.Context, userID, refToken string) (*domain.RefreshToken, error) {
	collection := tr.database.Collection(tr.collection)
	filter := bson.D{{Key: "user_id", Value: userID}}

	refreshToken, err := refreshToken(userID, refToken)
	if err != nil {
		return nil, err	
	}

	_, err = collection.ReplaceOne(ctx, filter, refreshToken)
	if err != nil {
		return nil, err	
	}

	return &domain.RefreshToken{
		Token:     refreshToken.Token,
		ExpiresAt: refreshToken.ExpiresAt,
	}, nil
}

func (tr *refreshTokenRepository) GetRefreshToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	collection := tr.database.Collection(tr.collection)
	filter := bson.D{{Key: "token", Value: token}}

	var refreshTokenDTO refreshTokenDTO
	err := collection.FindOne(ctx, filter).Decode(&refreshTokenDTO)
	if err != nil {
		if err != nil {
			return nil, err
		}
	}

	return &domain.RefreshToken{
		UserID: refreshTokenDTO.UserID.Hex(),
		Token: refreshTokenDTO.Token,
		ExpiresAt: refreshTokenDTO.ExpiresAt,
	}, err
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

type refreshTokenDTO struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	Token     string        `bson:"token"`
	UserID    bson.ObjectID `bson:"user_id"`
	ExpiresAt time.Time     `bson:"expires_at"`
	CreatedAt time.Time     `bson:"created_at"`
}

func refreshToken(userID, token string) (*refreshTokenDTO, error) {
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}
	return &refreshTokenDTO{
		Token:     token,
		UserID:    oid,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // Example expiration time
		CreatedAt: time.Now(),
	}, nil
}