package domain

import (
	"context"
	"time"
)

type Blog struct {
	ID        string
	Title     string
	Content   string
	AuthorID  string
	Tags      []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Comment struct {
    ID        string
    BlogID    string
    AuthorID  string
    Content   string             
    CreatedAt time.Time          
}

type BlogMetrics struct {
	ID           string
	BlogID       string

    ViewCount    int                
    LikeCount    int                
    DislikeCount int                
    CommentCount int                
}

type ReactionType string

const (
	Like 	ReactionType = "like"
	Dislike ReactionType = "dislike"
)

type Reaction struct {
    ID        string
    BlogID    string
    UserID    string
    Type      ReactionType             
    CreatedAt time.Time          
}

type IBlogRepository interface {
	// Blog CRUD
	CreateBlog(ctx context.Context, blog *Blog) (*Blog, error)
	GetBlogByID(ctx context.Context, id string) (*Blog, error)
	UpdateBlog(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteBlog(ctx context.Context, id string) error

	// Blog Listing
	ListBlogs(ctx context.Context, page, limit int) ([]*Blog, error)
	ListBlogsByAuthor(ctx context.Context, authorID string) ([]*Blog, error)
	SearchBlogs(ctx context.Context, query string) ([]*Blog, error)

	// Blog Metrics
	BlogMetricsInitializer(ctx context.Context, blogID string) error
	GetBlogMetrics(ctx context.Context, blogID string) (*BlogMetrics, error)
	IncrementViewCount(ctx context.Context, blogID string) error
}

type IReactionRepository interface {
	// Reactions
	AddReaction(ctx context.Context, reaction *Reaction) error
	RemoveReaction(ctx context.Context, blogID, userID string) error
	GetReactionsForBlog(ctx context.Context, blogID string) ([]*Reaction, error)
}

type ICommentRepository interface {
	// Comments
	AddComment(ctx context.Context, comment *Comment) (*Comment, error)
	GetCommentsForBlog(ctx context.Context, blogID string) ([]*Comment, error)
	DeleteComment(ctx context.Context, commentID string) error
}

type IBlogUseCase interface {
	CreateBlog(ctx context.Context, blog *Blog) (*Blog, error)
	GetBlog(ctx context.Context, blogID string) (*Blog, *BlogMetrics, error)
	UpdateBlog(ctx context.Context, blogID string, updates map[string]interface{}) error
	DeleteBlog(ctx context.Context, blogID, authorID string) error
	ListBlogs(ctx context.Context, page, limit int) ([]*Blog, error)
	SearchBlogs(ctx context.Context, query string) ([]*Blog, error)

	// Reactions
	AddReaction(ctx context.Context, blogID, userID string, reactionType string) error
	RemoveReaction(ctx context.Context, blogID, userID string) error

	// Comments
	AddComment(ctx context.Context, blogID, authorID string, content string) (*Comment, error)
	GetComments(ctx context.Context, blogID string) ([]*Comment, error)
}