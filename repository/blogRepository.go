package repository

import (
	"blog-backend/domain"
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type blogRepository struct {
	database   *mongo.Database
	collection string
}

type historyRepository struct {
	database   *mongo.Database
	collection string
}



func NewHistoryRepositoryFromDB(db *mongo.Database) domain.IHistoryRepository {
	return &historyRepository{
		database:   db,
		collection: "read_history",
	}
}
func NewBlogRepositoryFromDB(db *mongo.Database) domain.IBlogRepository {
	return &blogRepository{
		database:   db,
		collection: "blogs",
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
func (br *blogRepository) ListBlogs(ctx context.Context, page, limit int, field string) ([]*domain.Blog, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	collection := br.database.Collection(br.collection)
	skip := int64((page - 1) * limit)
	lim := int64(limit)
	findOptions := options.Find()
	findOptions.SetSkip(skip)
	findOptions.SetLimit(lim)
	findOptions.SetSort(bson.D{{Key: field, Value: -1}})

	cursor, err := collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, 0, err
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

func (br *blogRepository) UpdateBlogMetrics(ctx context.Context, blogID string, field string, reaction int) error {
	collection := br.database.Collection(br.collection)
	oid, err := bson.ObjectIDFromHex(blogID)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": oid}
	updates := bson.M{"$inc": bson.M{field: reaction}}
	res, err := collection.UpdateOne(ctx, filter, updates)
	if res.ModifiedCount == 0 {
		return fmt.Errorf("no blog found with id %s", blogID)
	}

	return err
}

// i added this function because we didn't have a function that evaluate blog authers k
func (br *blogRepository) IsAuthor(ctx context.Context, blogID, userID string) (bool, error) {
	collection := br.database.Collection(br.collection)
	oid, err := bson.ObjectIDFromHex(blogID)
	if err != nil {
		return false, err
	}
	ouid, err := bson.ObjectIDFromHex(userID)
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


func (h *historyRepository) AddReadHistory(ctx context.Context, userID, blogID string, blogTags []string) error {
    collection := h.database.Collection(h.collection)

    // get an UpdateOptions with Upsert=true using the helper
    upsertOpts := options.UpdateOne().SetUpsert(true)

    // 1. Add read entry (upsert to create history doc if missing)
    readUpdate := bson.M{
        "$push": bson.M{
            "reads": bson.M{
                "blog_id":    blogID,
                "created_at": time.Now(),
            },
        },
        "$setOnInsert": bson.M{
            "user_id": userID,
            "tags":    []interface{}{},
        },
    }
    _, err := collection.UpdateOne(ctx, bson.M{"user_id": userID}, readUpdate, upsertOpts)
    if err != nil {
        return fmt.Errorf("failed to update read history: %v", err)
    }

    // 2. Increment or insert tags
    for _, tag := range blogTags {
        filter := bson.M{"user_id": userID, "tags.tag": tag}
        update := bson.M{"$inc": bson.M{"tags.$.count": 1}}

        result, err := collection.UpdateOne(ctx, filter, update)
        if err != nil {
            return fmt.Errorf("failed to increment tag: %v", err)
        }

        // If the tag doesn't exist, push it (use upsert opts to be safe)
        if result.MatchedCount == 0 {
            _, err = collection.UpdateOne(
                ctx,
                bson.M{"user_id": userID},
                bson.M{"$push": bson.M{"tags": bson.M{"tag": tag, "count": 1}}},
                upsertOpts,
            )
            if err != nil {
                return fmt.Errorf("failed to add new tag: %v", err)
            }
        }
    }

    return nil
}


func (h *historyRepository) GetRecommendations(ctx context.Context, userID string) ([]*domain.Blog, error) {
    collection := h.database.Collection(h.collection)
    blogRepo := NewBlogRepositoryFromDB(h.database) // same DB assumed

    // Fetch only the tags array from the user's history
    var onlyTags struct {
        Tags []domain.TagsCount `bson:"tags"`
    }
    err := collection.FindOne(
        ctx,
        bson.M{"user_id": userID},
        options.FindOne().SetProjection(bson.M{"tags": 1}),
    ).Decode(&onlyTags)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return []*domain.Blog{}, nil // no history -> empty
        }
        return nil, fmt.Errorf("failed to get read history: %v", err)
    }

    tags := onlyTags.Tags
    if len(tags) == 0 {
        return []*domain.Blog{}, nil
    }

    // sort tags by count desc and take top 3
    sort.Slice(tags, func(i, j int) bool {
        return tags[i].Count > tags[j].Count
    })
    if len(tags) > 3 {
        tags = tags[:3]
    }

    // collect blogs for each top tag (deduplicate)
    blogSet := make(map[string]*domain.Blog)
    for _, t := range tags {
        blogs, err := blogRepo.SearchBlogs(ctx, t.Tag)
        if err != nil {
            return nil, fmt.Errorf("failed to search blogs for tag %s: %v", t.Tag, err)
        }
        for _, b := range blogs {
            blogSet[b.ID] = b
        }
    }

    // convert to slice and sort by view count desc
    blogs := make([]*domain.Blog, 0, len(blogSet))
    for _, b := range blogSet {
        blogs = append(blogs, b)
    }
    sort.Slice(blogs, func(i, j int) bool {
        return blogs[i].ViewCount > blogs[j].ViewCount
    })

    // limit to top 3
    if len(blogs) > 3 {
        blogs = blogs[:3]
    }

    return blogs, nil
}


type BlogResponseDTO struct {
	ID           bson.ObjectID `bson:"_id"`
	Title        string        `bson:"title" binding:"required"`
	Content      string        `bson:"content" binding:"required"`
	AuthorID     bson.ObjectID `bson:"author_id" binding:"required"`
	Tags         []string      `bson:"tags" binding:"required"`
	CreatedAt    time.Time     `bson:"created_at"`
	UpdatedAt    time.Time     `bson:"updated_at"`
	ViewCount    int           `bson:"view_count"`
	LikeCount    int           `bson:"like_count"`
	DislikeCount int           `bson:"dislike_count"`
	CommentCount int           `bson:"comment_count"`
}

type BlogDTO struct {
	Title        string        `bson:"title" binding:"required"`
	Content      string        `bson:"content" binding:"required"`
	AuthorID     bson.ObjectID `bson:"author_id" binding:"required"`
	Tags         []string      `bson:"tags" binding:"required"`
	CreatedAt    time.Time     `bson:"created_at"`
	UpdatedAt    time.Time     `bson:"updated_at"`
	ViewCount    int           `bson:"view_count"`
	LikeCount    int           `bson:"like_count"`
	DislikeCount int           `bson:"dislike_count"`
	CommentCount int           `bson:"comment_count"`
}

func DomainToDto(blog *domain.Blog) (*BlogDTO, error) {
	oid, err := bson.ObjectIDFromHex(blog.AuthorID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	return &BlogDTO{
		Title:        blog.Title,
		Content:      blog.Content,
		AuthorID:     oid,
		Tags:         blog.Tags,
		CreatedAt:    now,
		UpdatedAt:    now,
		ViewCount:    blog.ViewCount,
		LikeCount:    blog.LikeCount,
		DislikeCount: blog.DislikeCount,
		CommentCount: blog.CommentCount,
	}, err
}

func DtoToDomain(blogDTO *BlogResponseDTO) *domain.Blog {
	return &domain.Blog{
		ID:           blogDTO.ID.Hex(),
		Title:        blogDTO.Title,
		Content:      blogDTO.Content,
		AuthorID:     blogDTO.AuthorID.Hex(),
		Tags:         blogDTO.Tags,
		CreatedAt:    blogDTO.CreatedAt,
		UpdatedAt:    blogDTO.UpdatedAt,
		ViewCount:    blogDTO.ViewCount,
		LikeCount:    blogDTO.LikeCount,
		DislikeCount: blogDTO.DislikeCount,
		CommentCount: blogDTO.CommentCount,
	}
}

type HistoryDTO struct{
	UserID 		string   `bson: "user_id" binding: "required"` 
	BlogID		 string 	`bson: "blog_id" binding: "required"`
	CreatedAt time.Time		`bson:"created_at"`
	Tags     []domain.TagsCount	`bson: "tags"`
}
