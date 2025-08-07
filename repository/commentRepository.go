package repository

import (
	"blog-backend/domain"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
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
	collection := cr.database.Collection(cr.collection)
	_, err := collection.InsertOne(ctx, comment)
	if err != nil {
		return nil, err
	}
	return comment, nil
}

func (cr *commentRepository) GetCommentsForBlog(ctx context.Context, blogID string) ([]*domain.Comment, error) {
	collection := cr.database.Collection(cr.collection)
	
	filter := bson.M{"blogid": blogID}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var commentResDTO []CommentResDTO
	if err = cursor.All(ctx, &commentResDTO); err != nil {
		return nil, err
	}
	comments := make([]*domain.Comment, len(commentResDTO))
	for i, dto := range commentResDTO {
		comments[i] = CommentDtoToDomain(&dto)
	}
	return comments, nil
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
func (cr *commentRepository) IsComAuthor(ctx context.Context, comId, userId string) (bool,error) {
	collection := cr.database.Collection(cr.collection)
	oid, err := bson.ObjectIDFromHex(comId)
	if err != nil {
		return false, err
	}
	ouid,err := bson.ObjectIDFromHex(userId)
	if err != nil {
		return false, err
	}
	filter := bson.M{"_id": oid, "authorid": ouid}
	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
// type CommentResDTO struct {
// 	BlogID    string `bson:"blogid" json:"blogid"`
// 	AuthorID  string `bson:"authorid" json:"authorid"`
// 	Content   string `bson:"content" 	json:"content"`
// 	CreatedAt time.Time `bson:"created_at" json:"created_at"`
// }
//   func CommentDtoToDomain(dto *CommentResDTO) *domain.Comment {
// 	return &domain.Comment{
// 		BlogID:    dto.BlogID,
// 		AuthorID:  dto.AuthorID,
// 		Content:   dto.Content,
// 		CreatedAt: dto.CreatedAt,
// 	}
// }

