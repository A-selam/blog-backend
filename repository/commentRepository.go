package repository

import (
	"blog-backend/domain"
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type commentRepository struct {
	database   *mongo.Database
	collection string
}

func NewCommentRepositoryFromDB(db *mongo.Database) domain.ICommentRepository {
	return &commentRepository{
		database:   db,
		collection: "comments",
	}
}

func (cr *commentRepository) AddComment(ctx context.Context, comment *domain.Comment) (*domain.Comment, error) {
	// TODO: implement this function
	return nil, nil
}

func (cr *commentRepository) GetCommentsForBlog(ctx context.Context, blogID string) ([]*domain.Comment, error) {
	// TODO: implement this function
	return nil, nil
}

func (cr *commentRepository) DeleteComment(ctx context.Context, commentID string) error {
	collection := cr.database.Collection(cr.collection)
	oid, err := bson.ObjectIDFromHex(commentID)
	if err != nil {
		return err
	}
	_, err = collection.DeleteOne(ctx,bson.M{"_id":oid})
	return err
}
