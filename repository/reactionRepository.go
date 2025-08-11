package repository

import (
	"blog-backend/domain"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
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
	collection := rr.database.Collection(rr.collection)

	reactionDTO, err := domainToReactionDTO(reaction)
	if err != nil {
		return err
	}

	_, err = collection.InsertOne(ctx, reactionDTO)
	if err != nil {
		return err
	}

	return nil
}

func (rr *reactionRepository) RemoveReaction(ctx context.Context, blogID, userID string) error {
	collection := rr.database.Collection(rr.collection)

	bID, err := bson.ObjectIDFromHex(blogID)
	if err != nil {
		return err
	}

	uID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	
	filter := bson.D{{Key: "blog_id", Value: bID}, {Key: "user_id", Value: uID}}

	_, err = collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	return nil
}

func (rr *reactionRepository) CheckReactionExists(ctx context.Context, blogID, userID string) (*domain.Reaction, bool, error) {
	collection := rr.database.Collection(rr.collection)

	bID, err := bson.ObjectIDFromHex(blogID)
	if err != nil {
		return &domain.Reaction{}, false, err
	}

	uID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return &domain.Reaction{}, false, err
	}

	filter := bson.D{{Key: "blog_id", Value: bID}, {Key: "user_id", Value: uID}}

	var ReactionDTO ReactionDTO
	err = collection.FindOne(ctx, filter).Decode(&ReactionDTO)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &domain.Reaction{}, true, nil
		}

		return &domain.Reaction{}, false, err
	}

	dReaction, err := ReactionDTOToDomain(&ReactionDTO)
	if err != nil {
		return nil, false, err
	}

	return dReaction, false, nil
}

func (rr *reactionRepository) UpdateReaction(ctx context.Context, blogID, userID string, reactionType domain.ReactionType) error {
	collection := rr.database.Collection(rr.collection)
	
	bID, err := bson.ObjectIDFromHex(blogID)
	if err != nil {
		return err
	}
	uID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	filter := bson.D{{Key: "blog_id", Value: bID}, {Key: "user_id", Value: uID}}

	update := bson.D{{Key: "$set", Value: bson.D{{Key: "type", Value: string(reactionType)}}}}

	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}




type ReactionDTO struct {
	BlogID	    bson.ObjectID `bson:"blog_id"`
	UserID	    bson.ObjectID `bson:"user_id"`
	Type		string        `bson:"type"`	
	CreatedAt   time.Time     `bson:"created_at"`
}

func ReactionDTOToDomain(dto *ReactionDTO) (*domain.Reaction, error) {
	blogID := dto.BlogID.Hex()
	userID := dto.UserID.Hex()

	return &domain.Reaction{
		BlogID:    blogID,
		UserID:    userID,
		Type:      domain.ReactionType(dto.Type),
		CreatedAt: dto.CreatedAt,
	}, nil
}

func domainToReactionDTO(reaction *domain.Reaction) (*ReactionDTO, error) {
	blogID, err := bson.ObjectIDFromHex(reaction.BlogID)
	if err != nil {
		return nil, err
	}
	userID, err := bson.ObjectIDFromHex(reaction.UserID)
	if err != nil {
		return nil, err
	}
	return &ReactionDTO{
		BlogID:    blogID,
		UserID:    userID,
		Type:      string(reaction.Type),
		CreatedAt: reaction.CreatedAt,
	}, nil
}