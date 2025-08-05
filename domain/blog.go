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

type UpdateMetricsField string 

const (
	LikeCountField    UpdateMetricsField = "like_count"
	DislikeCountField UpdateMetricsField = "dislike_count"
)

type IBlogMetricsRepository interface {
	// Blog Metrics	
	BlogMetricsInitializer(ctx context.Context, blogID string) error
	GetBlogMetrics(ctx context.Context, blogID string) (*BlogMetrics, error)
	UpdateBlogMetrics(ctx context.Context, blogID string, field string, increment int) error
	IncrementViewCount(ctx context.Context, blogID string) error
}

type IBlogRepository interface {
	// Blog CRUD
	// IsAuthor(ctx context.Context, blogID, userID string) (bool, error)
	CreateBlog(ctx context.Context, blog *Blog) (*Blog, error)
	GetBlogByID(ctx context.Context, id string) (*Blog, error)
	UpdateBlog(ctx context.Context, blogID string, userID string, updates map[string]interface{}) error
	DeleteBlog(ctx context.Context, id string) error

	// Blog Listing
	ListBlogs(ctx context.Context, page, limit int) ([]*Blog, int64, error)
	ListBlogsByAuthor(ctx context.Context, authorID string) ([]*Blog, error)
	SearchBlogs(ctx context.Context, query string) ([]*Blog, error)

	//Blog authorization
	IsAuthor(ctx context.Context, blogID, userID string) (bool, error)
}

type IReactionRepository interface {
	// Reactions
	AddReaction(ctx context.Context, reaction *Reaction) error
	RemoveReaction(ctx context.Context, blogID, userID string) error
	GetReactionsForBlog(ctx context.Context, blogID string) ([]*Reaction, error)
	CheckReactionExists(ctx context.Context, blogID, userID string) (*Reaction, bool, error)
	UpdateReaction(ctx context.Context, blogID, userID string, reactionType ReactionType) error
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
	UpdateBlog(ctx context.Context, blogID string, userID string, updates map[string]interface{}) error
	DeleteBlog(ctx context.Context, blogID string) error
	ListBlogs(ctx context.Context, page, limit int) ([]*Blog,int64, error)
	SearchBlogs(ctx context.Context, query string) ([]*Blog, error)
	IsBlogAuthor(ctx context.Context, blogID, userID string) (bool, error)
	GetBlogsByUserID(ctx context.Context, userID string) ([]*Blog, error)
	GetBlogMetrics(ctx context.Context, blogID string) (*BlogMetrics, error)
	
	// Reactions
	AddReaction(ctx context.Context, blogID, userID string, reactionType string) error
	RemoveReaction(ctx context.Context, blogID, userID string) error

	// Comments
	AddComment(ctx context.Context, blogID, authorID string, content string) (*Comment, error)
	GetComments(ctx context.Context, blogID string) ([]*Comment, error)
	RemoveComment(ctx context.Context,commentID string)(error)
}