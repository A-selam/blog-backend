package usecase

import (
	"blog-backend/domain"
	"context"
	"time"
)

type blogUsecase struct {
	blogRepository domain.IBlogRepository
	blogReactionRepository domain.IReactionRepository
	blogCommentRepository domain.ICommentRepository
	contextTimeout time.Duration
}

func NewBlogUsecase(
	blogRepository        domain.IBlogRepository, 
	blogReactionRepository domain.IReactionRepository,
	blogCommentRepository domain.ICommentRepository,
	timeout time.Duration,
) domain.IBlogUseCase {
	return &blogUsecase{
		blogRepository: blogRepository,
		blogReactionRepository: blogReactionRepository,
		blogCommentRepository: blogCommentRepository,
		contextTimeout: timeout,
	}
}

func (bu *blogUsecase) CreateBlog(ctx context.Context, blog *domain.Blog) (*domain.Blog, error) {
	ctx, cancel := context.WithTimeout(ctx, bu.contextTimeout)
	defer cancel()

	createdBlog, err := bu.blogRepository.CreateBlog(ctx, blog)
	if err != nil {	
		// fmt.Println(err)
		return nil, err
	}	
		
	// Initialize blog metrics
	err = bu.blogRepository.BlogMetricsInitializer(ctx, createdBlog.ID)
	if err != nil {	
		return nil, err
	}

	return createdBlog, nil
}

func (bu *blogUsecase) GetBlog(ctx context.Context, blogID string) (*domain.Blog, *domain.BlogMetrics, error) {
	// TODO: implement this function
	return nil, nil, nil
}

func (bu *blogUsecase) UpdateBlog(ctx context.Context, blogID string, updates map[string]interface{}) error {
	// TODO: implement this function
	return nil
}

func (bu *blogUsecase) DeleteBlog(ctx context.Context, blogID, authorID string) error {
	// TODO: implement this function
	return nil
}

func (bu *blogUsecase) ListBlogs(ctx context.Context, page, limit int) ([]*domain.Blog, error) {
	// TODO: implement this function
	return nil, nil
}

func (bu *blogUsecase) SearchBlogs(ctx context.Context, query string) ([]*domain.Blog, error) {
	// TODO: implement this function
	return nil, nil
}

// Reactions
func (bu *blogUsecase) AddReaction(ctx context.Context, blogID, userID string, reactionType string) error {
	// TODO: implement this function
	return nil
}

func (bu *blogUsecase) RemoveReaction(ctx context.Context, blogID, userID string) error {
	// TODO: implement this function
	return nil
}

// Comments
func (bu *blogUsecase) AddComment(ctx context.Context, blogID, authorID string, content string) (*domain.Comment, error) {
	// TODO: implement this function
	return nil, nil
}

func (bu *blogUsecase) GetComments(ctx context.Context, blogID string) ([]*domain.Comment, error) {
	// TODO: implement this function
	return nil, nil
}
