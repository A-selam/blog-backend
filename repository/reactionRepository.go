package repository

import (
	"blog-backend/domain"
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type reactionRepository struct {
	database   *mongo.Database
	collection string
}

func NewReactionRepositoryFromDB(db *mongo.Database) domain.IReactionRepository {
	return &reactionRepository{
		database:   db,
		collection: "reactions",
	}
}

func (rr *reactionRepository) AddReaction(ctx context.Context, reaction *domain.Reaction) error {
	// TODO: implement this function
	return nil
}

func (rr *reactionRepository) RemoveReaction(ctx context.Context, blogID, userID string) error {
	// TODO: implement this function
	return nil
}

func (rr *reactionRepository) GetReactionsForBlog(ctx context.Context, blogID string) ([]*domain.Reaction, error) {
	// TODO: implement this function
	return nil, nil
}
