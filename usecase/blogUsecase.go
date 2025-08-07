package usecase

import (
	"blog-backend/domain"
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type blogUsecase struct {
	blogRepository         domain.IBlogRepository
	blogReactionRepository domain.IReactionRepository
	blogCommentRepository  domain.ICommentRepository
	geminiServices         domain.IGeminiService
	contextTimeout         time.Duration
}



func NewBlogUsecase(
	blogRepository domain.IBlogRepository,
	blogReactionRepository domain.IReactionRepository,
	blogCommentRepository domain.ICommentRepository,
	geminiServices domain.IGeminiService,
	timeout time.Duration,
) domain.IBlogUseCase {
	return &blogUsecase{
		blogRepository:         blogRepository,
		blogReactionRepository: blogReactionRepository,
		blogCommentRepository:  blogCommentRepository,
		geminiServices:         geminiServices,
		contextTimeout:         timeout,
	}
}

func (bu *blogUsecase) CreateBlog(ctx context.Context, blog *domain.Blog) (*domain.Blog, error) {
	ctx, cancel := context.WithTimeout(ctx, bu.contextTimeout)
	defer cancel()

	if len(blog.Tags) == 0 {
		fullPrompt := fmt.Sprintf(`Analyze the following blog post content and generate 5 relevant tags in a comma-separated list: "%s"`, blog.Content)

		response, err := bu.geminiServices.GenerateTags(fullPrompt)
		if err != nil {
			return nil, err
		}

		blog.Tags = response
	}

	createdBlog, err := bu.blogRepository.CreateBlog(ctx, blog)
	if err != nil {
		return nil, err
	}

	return createdBlog, nil
}

func (bu *blogUsecase) GetBlog(ctx context.Context, blogID string) (*domain.Blog, error) {
	ctx, cancel := context.WithTimeout(ctx, bu.contextTimeout)
	defer cancel()
	blog, err := bu.blogRepository.GetBlogByID(ctx, blogID)
	if err != nil {
		return nil, err
	}

	errChan := make(chan error, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func ()  {
		defer wg.Done()
		goroutineCtx, goroutineCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer goroutineCancel()
	err = bu.blogRepository.UpdateBlogMetrics(goroutineCtx, blogID, "view_count", 1)
	if err != nil {		errChan <- err
		log.Printf("Error updating view count for blog %s: %v", blogID, err)
		return

	}
	}()
	go func() {
	wg.Wait()
	close(errChan)
	}()
	for err := range errChan {
		if err != nil {	
			return nil, err
		}
	
	}
	return blog, nil
}

func (bu *blogUsecase) UpdateBlog(ctx context.Context, blogID string, userID string, updates map[string]interface{}) error {
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

func (bu *blogUsecase) ListBlogs(ctx context.Context, page, limit int, field string) ([]*domain.Blog, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, bu.contextTimeout)
	defer cancel()
	blogs, total, err := bu.blogRepository.ListBlogs(ctx, page, limit, field)
	if err != nil {
		return nil, 0, err
	}
	return blogs, total, nil
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

func (bu *blogUsecase) AddReaction(ctx context.Context, blogID, userID string, reactionType string) error {
	if _, err := bson.ObjectIDFromHex(blogID); err != nil {
		return errors.New("invalid blog ID")
	}

	ctx, cancel := context.WithTimeout(ctx, bu.contextTimeout)
	defer cancel()

	reaction := &domain.Reaction{
		BlogID:    blogID,
		UserID:    userID,
		Type:      domain.ReactionType(reactionType),
		CreatedAt: time.Now(),
	}
	rxn, noReaction, err := bu.blogReactionRepository.CheckReactionExists(ctx, blogID, userID)
	if err != nil {
		log.Printf("Error checking reaction: %v", err)
		return err
	}

	goroutineCtx, goroutineCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer goroutineCancel()

	// Channel to collect errors from goroutines
	errChan := make(chan error, 1)
	var wg sync.WaitGroup

	if noReaction {
		err = bu.blogReactionRepository.AddReaction(ctx, reaction)
		if err != nil {
			log.Printf("Error adding reaction: %v", err)
			return err
		}

		// Asynchronously update metrics
		wg.Add(1)
		go func() {
			defer wg.Done()
			var field string
			if reactionType == string(domain.Like) {
				field = string(domain.LikeCountField)
			} else {
				field = string(domain.DislikeCountField)
			}
			if err := bu.blogRepository.UpdateBlogMetrics(goroutineCtx, blogID, field, 1); err != nil {
				log.Printf("Failed to update %s for blog %s: %v", field, blogID, err)
				errChan <- fmt.Errorf("failed to update %s: %v", field, err)
			}
		}()
	} else if string(rxn.Type) != reactionType {
		err = bu.blogReactionRepository.UpdateReaction(ctx, blogID, userID, domain.ReactionType(reactionType))
		if err != nil {
			log.Printf("Error updating reaction: %v", err)
			return err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			if reactionType == string(domain.Like) {
				if err := bu.blogRepository.UpdateBlogMetrics(goroutineCtx, blogID, string(domain.LikeCountField), 1); err != nil {
					log.Printf("Failed to increment like count for blog %s: %v", blogID, err)
					errChan <- fmt.Errorf("failed to increment like count: %v", err)
				}
				if err := bu.blogRepository.UpdateBlogMetrics(goroutineCtx, blogID, string(domain.DislikeCountField), -1); err != nil {
					log.Printf("Failed to decrement dislike count for blog %s: %v", blogID, err)
					errChan <- fmt.Errorf("failed to decrement dislike count: %v", err)
				}
			} else if reactionType == string(domain.Dislike) {
				if err := bu.blogRepository.UpdateBlogMetrics(goroutineCtx, blogID, string(domain.DislikeCountField), 1); err != nil {
					log.Printf("Failed to increment dislike count for blog %s: %v", blogID, err)
					errChan <- fmt.Errorf("failed to increment dislike count: %v", err)
				}
				if err := bu.blogRepository.UpdateBlogMetrics(goroutineCtx, blogID, string(domain.LikeCountField), -1); err != nil {
					log.Printf("Failed to decrement like count for blog %s: %v", blogID, err)
					errChan <- fmt.Errorf("failed to decrement like count: %v", err)
				}
			}
		}()
	} else {
		return errors.New("reaction already exists with the same type")
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		if err != nil {

			return err
		}
	}

	return nil
}

func (bu *blogUsecase) RemoveReaction(ctx context.Context, blogID, userID string) error {
	ctx, cancel := context.WithTimeout(ctx, bu.contextTimeout)
	defer cancel()
	rxn, noReaction, err := bu.blogReactionRepository.CheckReactionExists(ctx, blogID, userID)
	if err != nil {
		return err
	}
	err = bu.blogReactionRepository.RemoveReaction(ctx, blogID, userID)
	if err != nil {
		return err
	}
	if noReaction {
		return errors.New("no reaction found to remove")
	}
	errChan := make(chan error, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func(){
		defer wg.Done()
		goroutineCtx, goroutineCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer goroutineCancel()
	if string(rxn.Type) == string(domain.Like) {
		err = bu.blogRepository.UpdateBlogMetrics(goroutineCtx, blogID, string(domain.LikeCountField), -1)
		if err != nil {
			errChan <- err
		}
	} else if rxn.Type == domain.Dislike {
		err = bu.blogRepository.UpdateBlogMetrics(goroutineCtx, blogID, string(domain.DislikeCountField), -1)
		if err != nil {
			errChan <- err
		}
	}else {
		 errChan <- errors.New("reaction type not recognized")
	}

	}()
	go func() {
	wg.Wait()
	close(errChan)

	}()
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// Comments
func (bu *blogUsecase) AddComment(ctx context.Context, comment *domain.Comment) (*domain.Comment, error) {
	ctx, cancel := context.WithTimeout(ctx, bu.contextTimeout)
	defer cancel()
	
	res, err := bu.blogCommentRepository.AddComment(ctx, comment)
	if err != nil {
		return nil, err
	}
	errcChan := make(chan error, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		goroutineCtx, goroutineCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer goroutineCancel()
	err = bu.blogRepository.UpdateBlogMetrics(goroutineCtx, comment.BlogID, "comment_count", 1)
	if err != nil {
		log.Printf("Error updating comment count for blog %s: %v", comment.BlogID, err)
		errcChan <- fmt.Errorf("failed to update comment count: %v", err)
		return
	}
}()
	go func() {
	wg.Wait()
	close(errcChan)
	}()
	for err := range errcChan {
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}
func (bu *blogUsecase) IsComAuthor(ctx context.Context, comId, userId string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, bu.contextTimeout)
	defer cancel()
	return bu.blogCommentRepository.IsComAuthor(ctx, comId, userId)
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

func (bu *blogUsecase) RemoveComment(ctx context.Context,commentID string)(error){
	ctx, cancel := context.WithTimeout(ctx,bu.contextTimeout)
	defer cancel()
	return bu.blogCommentRepository.DeleteComment(ctx,commentID)
}
