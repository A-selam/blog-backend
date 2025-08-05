package repository

import (
	"blog-backend/domain"
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type blogMetricsRepository struct {
	database   *mongo.Database
	collection string
}


type blogRepository struct {
	database   *mongo.Database
	collection string
}

func NewBlogRepositoryFromDB(db *mongo.Database) domain.IBlogRepository {
	return &blogRepository{
		database:   db,
		collection: "blogs",
	}
}

func NewBlogMetricsRepositoryFromDB(db *mongo.Database) domain.IBlogMetricsRepository {
	return &blogMetricsRepository{	
		database:   db,
		collection: "blog_metrics",
	}
}
// Blog CRUD
func (br *blogRepository) CreateBlog(ctx context.Context, blog *domain.Blog) (*domain.Blog, error) {
	collection := br.database.Collection(br.collection)

	blogDTO, err := DomainToDto(blog)
	if err != nil {
		return nil, err
	}

	insertedResult, err := collection.InsertOne(ctx, blogDTO)
	if err != nil {
		return nil, err
	}
	id := insertedResult.InsertedID.(bson.ObjectID)
	blog.ID = id.Hex()

	return blog, nil
}

func (bmr *blogMetricsRepository) BlogMetricsInitializer(ctx context.Context, blogID string) error {
	collection := bmr.database.Collection(bmr.collection)

	id, err := bson.ObjectIDFromHex(blogID)
	if err != nil {
		return err
	}

	blogMetrics := BlogMetricsDTO{
		BlogID:       id,
		ViewCount:    0,
		LikeCount:    0,
		DislikeCount: 0,
		CommentCount: 0,
	}

	_, err = collection.InsertOne(ctx, blogMetrics)
	return err
}

func (br *blogRepository) GetBlogByID(ctx context.Context, id string) (*domain.Blog, error) {
	collection := br.database.Collection(br.collection)
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var blog BlogResponseDTO
	filter := bson.M{"_id": oid}
	err = collection.FindOne(ctx, filter).Decode(&blog)
	if err != nil {
		return nil, err
	}

	return DtoToDomain(&blog), nil
}
func (br *blogRepository) UpdateBlog(ctx context.Context, id string, userID string, updates map[string]interface{}) error {
    collection := br.database.Collection(br.collection)
    oid, err := bson.ObjectIDFromHex(id)
    if err != nil {
        return err
    }
    var blog BlogResponseDTO
    filter := bson.M{"_id": oid}
    err = collection.FindOne(ctx, filter).Decode(&blog)
    if err != nil {
        return err
    }

    authorID := blog.AuthorID.Hex()
    if authorID != userID {
        
        userRepo := NewUserRepositoryFromDB(br.database)
        user, err := userRepo.GetUserByID(ctx, userID)
        if err != nil {
			log.Print("unauthorized: user not found")
            return fmt.Errorf("unauthorized: user not found")
        }
        if user.Role != domain.Admin {
			log.Print("unauthorized: not author or admin")

            return fmt.Errorf("unauthorized: not author or admin")
        }
    }

    update := bson.M{"$set": updates}
    _, err = collection.UpdateOne(ctx, filter, update)
    return err
}
func (br *blogRepository) DeleteBlog(ctx context.Context, id string) error {
	collection := br.database.Collection(br.collection)
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = collection.DeleteOne(ctx, bson.M{"_id": oid})
	return err
}

// Blog Listing
func (br *blogRepository) ListBlogs(ctx context.Context, page, limit int) ([]*domain.Blog,int64, error) {
	collection := br.database.Collection(br.collection)
	skip := int64((page - 1) * limit)
	lim := int64(limit)
	findOptions := options.Find()
	findOptions.SetSkip(skip)
	findOptions.SetLimit(lim)
	cursor, err := collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil,0, err
	}
	defer cursor.Close(ctx)
	var blogResDTOs []BlogResponseDTO
	if err = cursor.All(ctx, &blogResDTOs); err != nil {
		return nil, 0, err
	}
	blogs := make([]*domain.Blog, len(blogResDTOs))
	for i, dto := range blogResDTOs {
		blogs[i] = DtoToDomain(&dto)
	}
	total, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}

	return blogs, total, nil
}
func (br *blogRepository) ListBlogsByAuthor(ctx context.Context, authorID string) ([]*domain.Blog, error) {
	collection := br.database.Collection(br.collection)
	filter := bson.M{"author_id": authorID}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var blogResDTOs []BlogResponseDTO
	if err = cursor.All(ctx, &blogResDTOs); err != nil {
		return nil, err
	}

	blogs := make([]*domain.Blog, len(blogResDTOs))
	for i, dto := range blogResDTOs {
		blogs[i] = DtoToDomain(&dto)
	}

	return blogs, nil
}
func (br *blogRepository) SearchBlogs(ctx context.Context, query string) ([]*domain.Blog, error) {
	collection := br.database.Collection(br.collection)
	filter := bson.M{
		"$or": []bson.M{
			{"title": bson.M{"$regex": query, "$options": "i"}},
			{"tags": bson.M{"$regex": query, "$options": "i"}},
		},
	}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var blogResDTOs []BlogResponseDTO
	if err = cursor.All(ctx, &blogResDTOs); err != nil {
		return nil, err
	}

	blogs := make([]*domain.Blog, len(blogResDTOs))
	for i, dto := range blogResDTOs {
		blogs[i] = DtoToDomain(&dto)
	}

	return blogs, nil
}

// Blog Metrics
func (bmr *blogMetricsRepository) GetBlogMetrics(ctx context.Context, blogID string) (*domain.BlogMetrics, error) {
	collection := bmr.database.Collection(bmr.collection)
	oid, err := bson.ObjectIDFromHex(blogID)
	if err != nil{
		return nil, err
	}
	filter := bson.M{"blog_id": oid}
	var blog_metrics BlogMetricsDTO
	err = collection.FindOne(ctx, filter).Decode(&blog_metrics)
	if err != nil {
		return nil, err
	}

	return BlogMetricsDtoToDomain(&blog_metrics), nil
}

func (bmr *blogMetricsRepository) UpdateBlogMetrics(ctx context.Context, blogID string, field string, reaction int) error {
	collection := bmr.database.Collection(bmr.collection)
	oid, err := bson.ObjectIDFromHex(blogID)
	if err != nil {
		return err
	}
	filter := bson.M{"blog_id": oid}
	updates := bson.M{"$inc": bson.M{field: reaction}}
	_, err = collection.UpdateOne(ctx, filter, updates)
	return err
}

func (bmr *blogMetricsRepository) IncrementViewCount(ctx context.Context, blogID string) error {
	// TODO: implement this function
	return nil
}

func (br *blogRepository) IncrementLikeCount(ctx context.Context, blogID string) error {
	collection := br.database.Collection("blog_metrics")
	oid, err := bson.ObjectIDFromHex(blogID)
	if err != nil {
		return err
	}

	filter := bson.D{{Key: "blog_id", Value: oid}}
	update := bson.D{{Key: "$inc", Value: bson.D{{Key: "like_count", Value: 1}}}}
	
	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (br *blogRepository) DecrementLikeCount(ctx context.Context, blogID string) error {
	collection := br.database.Collection("blog_metrics")
	oid, err := bson.ObjectIDFromHex(blogID)
	if err != nil {
		return err
	}

	filter := bson.D{{Key: "blog_id", Value: oid}}
	update := bson.D{{Key: "$dec", Value: bson.D{{Key: "like_count", Value: 1}}}}
	
	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (br *blogRepository) IncrementDislikeCount(ctx context.Context, blogID string) error {
	collection := br.database.Collection("blog_metrics")
	oid, err := bson.ObjectIDFromHex(blogID)
	if err != nil {
		return err
	}

	filter := bson.D{{Key: "blog_id", Value: oid}}
	update := bson.D{{Key: "$inc", Value: bson.D{{Key: "like_count", Value: 1}}}}
	
	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (br *blogRepository) DecrementDislikeCount(ctx context.Context, blogID string) error {
	collection := br.database.Collection("blog_metrics")
	oid, err := bson.ObjectIDFromHex(blogID)
	if err != nil {
		return err
	}

	filter := bson.D{{Key: "blog_id", Value: oid}}
	update := bson.D{{Key: "$dec", Value: bson.D{{Key: "like_count", Value: 1}}}}
	
	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

// i added this function because we didn't have a function that evaluate blog authers k
func (br *blogRepository) IsAuthor(ctx context.Context, blogID, userID string) (bool, error) {
	collection := br.database.Collection(br.collection)
	oid, err := bson.ObjectIDFromHex(blogID)
	ouid,err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return false, err
	}
	filter := bson.M{"_id": oid, "author_id": ouid}
	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}


type BlogMetricsDTO struct {
	BlogID       bson.ObjectID `bson:"blog_id"`
	ViewCount    int           `bson:"view_count"`
	LikeCount    int           `bson:"like_count"`
	DislikeCount int           `bson:"dislike_count"`
	CommentCount int           `bson:"comment_count"`
}

type BlogResponseDTO struct {
	ID        bson.ObjectID `bson:"_id"`
	Title     string        `bson:"title" binding:"required"`
	Content   string        `bson:"content" binding:"required"`
	AuthorID  bson.ObjectID `bson:"author_id" binding:"required"`
	Tags      []string      `bson:"tags" binding:"required"`
	CreatedAt time.Time     `bson:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at"`
}

type BlogDTO struct {
	Title     string        `bson:"title" binding:"required"`
	Content   string        `bson:"content" binding:"required"`
	AuthorID  bson.ObjectID `bson:"author_id" binding:"required"`
	Tags      []string      `bson:"tags" binding:"required"`
	CreatedAt time.Time     `bson:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at"`
}

func DomainToDto(blog *domain.Blog) (*BlogDTO, error) {
	oid, err := bson.ObjectIDFromHex(blog.AuthorID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	return &BlogDTO{
		Title:     blog.Title,
		Content:   blog.Content,
		AuthorID:  oid,
		Tags:      blog.Tags,
		CreatedAt: now,
		UpdatedAt: now,
	}, err
}

func DtoToDomain(blogDTO *BlogResponseDTO) *domain.Blog {
	return &domain.Blog{
		ID:        blogDTO.ID.Hex(),
		Title:     blogDTO.Title,
		Content:   blogDTO.Content,
		AuthorID:  blogDTO.AuthorID.Hex(),
		Tags:      blogDTO.Tags,
		CreatedAt: blogDTO.CreatedAt,
		UpdatedAt: blogDTO.UpdatedAt,
	}
}

func BlogMetricsDtoToDomain(dto *BlogMetricsDTO) *domain.BlogMetrics {
	return &domain.BlogMetrics{
		BlogID:       dto.BlogID.Hex(),
		ViewCount:    dto.ViewCount,
		LikeCount:    dto.LikeCount,
		DislikeCount: dto.DislikeCount,
		CommentCount: dto.CommentCount,
	}
}
