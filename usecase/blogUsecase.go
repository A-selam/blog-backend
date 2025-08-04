package usecase

import (
	"blog-backend/domain"
	"context"
	"time"
)

type blogUsecase struct {
	blogRepository         domain.IBlogRepository
	blogReactionRepository domain.IReactionRepository
	blogCommentRepository  domain.ICommentRepository
	blogMetricsRepository  domain.IBlogMetricsRepository
	contextTimeout         time.Duration
}

func NewBlogUsecase(
	blogRepository domain.IBlogRepository,
	blogReactionRepository domain.IReactionRepository,
	blogCommentRepository domain.ICommentRepository,
	timeout time.Duration,
	blogMetricsRepository domain.IBlogMetricsRepository,
) domain.IBlogUseCase {
	return &blogUsecase{
		blogMetricsRepository: blogMetricsRepository,
		blogRepository:         blogRepository,
		blogReactionRepository: blogReactionRepository,
		blogCommentRepository:  blogCommentRepository,
		contextTimeout:         timeout,
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
	err = bu.blogMetricsRepository.BlogMetricsInitializer(ctx, createdBlog.ID)
	if err != nil {
		return nil, err
	}

	return createdBlog, nil
}

func (bu *blogUsecase) GetBlog(ctx context.Context, blogID string) (*domain.Blog, *domain.BlogMetrics, error) {
	ctx, cancel := context.WithTimeout(ctx, bu.contextTimeout)
	defer cancel()
	blog, err := bu.blogRepository.GetBlogByID(ctx, blogID)
	if err != nil{
		return nil, nil, err
	}
	metric, err := bu.blogMetricsRepository.GetBlogMetrics(ctx, blogID)
	if err != nil{
		return nil, nil, err
	}

	return blog, metric, nil
}

func (bu *blogUsecase) UpdateBlog(ctx context.Context, blogID string,userID string,  updates map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, bu.contextTimeout)
	defer cancel()
	err := bu.blogRepository.UpdateBlog(ctx, blogID, userID, updates)
	return err
}

func (bu *blogUsecase) DeleteBlog(ctx context.Context, blogID string) error {
	ctx, cancel := context.WithTimeout(ctx, bu.contextTimeout)
	defer cancel()
	err := bu.blogRepository.DeleteBlog(ctx, blogID)
	return err
}

func (bu *blogUsecase) ListBlogs(ctx context.Context, page, limit int) ([]*domain.Blog,int64, error) {
	ctx, cancel := context.WithTimeout(ctx, bu.contextTimeout)
	defer cancel()
	blogs, total, err := bu.blogRepository.ListBlogs(ctx, page, limit)
	if err != nil {
		return nil,0, err
	}
	return blogs,total, nil
}

func (bu *blogUsecase) SearchBlogs(ctx context.Context, query string) ([]*domain.Blog, error) {
	ctx, cancel := context.WithTimeout(ctx, bu.contextTimeout)
	defer cancel()

	blogs, err := bu.blogRepository.SearchBlogs(ctx, query)
	if err != nil {
		return nil, err
	}
	return blogs, nil
}

// Reactions
func (bu *blogUsecase) AddReaction(ctx context.Context, blogID, userID string, reactionType string) error {
	// TODO: implement this function
	return nil
}

func (bu *blogUsecase) RemoveReaction(ctx context.Context, blogID, userID string) error {
	ctx, cancel := context.WithTimeout(ctx, bu.contextTimeout)
	defer cancel()
	err := bu.blogReactionRepository.RemoveReaction(ctx, blogID, userID)
	if err != nil {
		return err
	}
	err = bu.blogMetricsRepository.UpdateBlogMetrics(ctx, blogID, "reactions", -1)
	return err
}

// Comments
func (bu *blogUsecase) AddComment(ctx context.Context, blogID, authorID string, content string) (*domain.Comment, error) {
	ctx, cancel := context.WithTimeout(ctx, bu.contextTimeout)
	defer cancel()
	comment := &domain.Comment{
		BlogID:    blogID,
		AuthorID:  authorID,
		Content:   content,
		CreatedAt: time.Now(),
	}

	res, err := bu.blogCommentRepository.AddComment(ctx, comment)
	if err != nil{
		return nil, err
	}
	return res, nil
}

func (bu *blogUsecase) GetComments(ctx context.Context, blogID string) ([]*domain.Comment, error) {
	ctx, cancel := context.WithTimeout(ctx, bu.contextTimeout)
	defer cancel()
	comments, err := bu.blogCommentRepository.GetCommentsForBlog(ctx, blogID)
	if err != nil {
		return nil, err
	}
	return comments, nil
}
func (bu *blogUsecase) GetBlogsByUserID(ctx context.Context, userID string) ([]*domain.Blog, error) {
	ctx, cancel := context.WithTimeout(ctx, bu.contextTimeout)
	defer cancel()

	blogs, err := bu.blogRepository.ListBlogsByAuthor(ctx, userID)
	if err != nil {
		return nil, err
	}
	return blogs, nil
}
func (bu *blogUsecase) IsBlogAuthor(ctx context.Context, blogID, userID string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, bu.contextTimeout)
	defer cancel()

	isAuthor, err := bu.blogRepository.IsAuthor(ctx, blogID, userID)
	if err != nil {
		return false, err
	}
	return isAuthor, nil
}
