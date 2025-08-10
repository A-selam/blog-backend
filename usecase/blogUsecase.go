package usecase

import (
	"blog-backend/domain"
	"context"
	"encoding/json" 
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

//this are cache expiration durations
const (
	blogDetailCacheTTL = 10 * time.Minute // TTL for individual blog posts
	blogListCacheTTL   = 5 * time.Minute  // TTL for lists of blogs (all, by user, search)
	commentListCacheTTL = 2 * time.Minute // TTL for comments list
)

type blogUsecase struct {
	blogRepository         domain.IBlogRepository
	blogReactionRepository domain.IReactionRepository
	blogCommentRepository  domain.ICommentRepository
	historyRepository	domain.IHistoryRepository
	geminiServices         domain.IGeminiService
	cacheUseCase           domain.ICacheUseCase
	contextTimeout         time.Duration
}

func NewBlogUsecase(
	blogRepository domain.IBlogRepository,
	blogReactionRepository domain.IReactionRepository,
	blogCommentRepository domain.ICommentRepository,
	historyRepository	domain.IHistoryRepository,
	geminiServices domain.IGeminiService,
	timeout time.Duration,
	cacheUseCase domain.ICacheUseCase, 
) domain.IBlogUseCase {
	return &blogUsecase{
		blogRepository:         blogRepository,
		blogReactionRepository: blogReactionRepository,
		blogCommentRepository:  blogCommentRepository,
		historyRepository: 		historyRepository,	
		geminiServices:         geminiServices,
		contextTimeout:         timeout,
		cacheUseCase:           cacheUseCase, 
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

	go bu.cacheUseCase.InvalidatePrefix(context.Background(), "blogs:list:")
	go bu.cacheUseCase.InvalidatePrefix(context.Background(), fmt.Sprintf("blogs:user:%s", blog.AuthorID))

	return createdBlog, nil
}

func (bu *blogUsecase) 	AddReadHistory(ctx context.Context, userID, blogID string) error{
	ctx, cancel := context.WithTimeout(ctx, bu.contextTimeout)
	defer cancel()
	blog, err := bu.blogRepository.GetBlogByID(ctx, blogID)
	if err != nil{
		return err
	}

	err = bu.historyRepository.AddReadHistory(ctx, userID, blogID, blog.Tags)
	if err != nil{
		return err
	}
	return nil
}

func (bu *blogUsecase) 	GetRecommendations(ctx context.Context, userID string) ([]*domain.Blog, error){
	ctx, cancel := context.WithTimeout(ctx, bu.contextTimeout)
	defer cancel()
	blogs, err := bu.historyRepository.GetRecommendations(ctx, userID)
	if err != nil{
		return nil, err
	}
	return blogs, nil
	
}

func (bu *blogUsecase) GetBlog(ctx context.Context, blogID string) (*domain.Blog, error) {
	ctx, cancel := context.WithTimeout(ctx, bu.contextTimeout)
	defer cancel()

	cacheKey := fmt.Sprintf("blog:id:%s", blogID)

	cachedBlogBytes, err := bu.cacheUseCase.Get(ctx, cacheKey)
	if err == nil && cachedBlogBytes != nil {
		var blog domain.Blog
		if err := json.Unmarshal(cachedBlogBytes, &blog); err == nil {
			go func() {
				goroutineCtx, goroutineCancel := context.WithTimeout(context.Background(), 5*time.Second) 
				defer goroutineCancel()
				if updateErr := bu.blogRepository.UpdateBlogMetrics(goroutineCtx, blogID, "view_count", 1); updateErr != nil {
					log.Printf("Error updating view count for blog %s: %v", blogID, updateErr)
				} else {
					bu.cacheUseCase.Delete(goroutineCtx, cacheKey)
				}
			}()
			return &blog, nil
		}
		log.Printf("Failed to unmarshal cached blog %s: %v", blogID, err)
	} else if err != nil {
		log.Printf("Error getting blog from cache %s: %v", blogID, err)
	}

	blog, err := bu.blogRepository.GetBlogByID(ctx, blogID)
	if err != nil {
		return nil, err
	}

	blogJSON, err := json.Marshal(blog)
	if err == nil {
		bu.cacheUseCase.Set(ctx, cacheKey, blogJSON, blogDetailCacheTTL)
	} else {
		log.Printf("Failed to marshal blog %s for caching: %v", blogID, err)
	}

	go func() {
		goroutineCtx, goroutineCancel := context.WithTimeout(context.Background(), 5*time.Second) 
		defer goroutineCancel()
		if updateErr := bu.blogRepository.UpdateBlogMetrics(goroutineCtx, blogID, "view_count", 1); updateErr != nil {
			log.Printf("Error updating view count for blog %s: %v", blogID, updateErr)
		} else {
			bu.cacheUseCase.Delete(goroutineCtx, cacheKey)
		}
	}()

	return blog, nil
}

func (bu *blogUsecase) UpdateBlog(ctx context.Context, blogID string, userID string, updates map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, bu.contextTimeout)
	defer cancel()

	err := bu.blogRepository.UpdateBlog(ctx, blogID, userID, updates)
	if err != nil {
		return err
	}

	go bu.cacheUseCase.Delete(context.Background(), fmt.Sprintf("blog:id:%s", blogID))
	go bu.cacheUseCase.InvalidatePrefix(context.Background(), "blogs:list:")
	go bu.cacheUseCase.InvalidatePrefix(context.Background(), fmt.Sprintf("blogs:user:%s", userID)) 

	return nil
}

func (bu *blogUsecase) DeleteBlog(ctx context.Context, blogID string) error {
	ctx, cancel := context.WithTimeout(ctx, bu.contextTimeout)
	defer cancel()

	blog, err := bu.blogRepository.GetBlogByID(ctx, blogID)
	if err != nil { 
		return err
	}

	err = bu.blogRepository.DeleteBlog(ctx, blogID)
	if err != nil {
		return err
	}

	go bu.cacheUseCase.Delete(context.Background(), fmt.Sprintf("blog:id:%s", blogID))
	go bu.cacheUseCase.InvalidatePrefix(context.Background(), "blogs:list:")
	if blog != nil { 
		go bu.cacheUseCase.InvalidatePrefix(context.Background(), fmt.Sprintf("blogs:user:%s", blog.AuthorID))
	}

	return nil
}

func (bu *blogUsecase) ListBlogs(ctx context.Context, page, limit int, field string) ([]*domain.Blog, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, bu.contextTimeout)
	defer cancel()

	cacheKey := fmt.Sprintf("blogs:list:page:%d:limit:%d:field:%s", page, limit, field)

	cachedBlogsBytes, err := bu.cacheUseCase.Get(ctx, cacheKey)
	if err == nil && cachedBlogsBytes != nil {
		var cachedData struct {
			Blogs []*domain.Blog `json:"blogs"`
			Total int64          `json:"total"`
		}
		if err := json.Unmarshal(cachedBlogsBytes, &cachedData); err == nil {
			return cachedData.Blogs, cachedData.Total, nil
		}
		log.Printf("Failed to unmarshal cached blog list %s: %v", cacheKey, err)
	} else if err != nil {
		log.Printf("Error getting blog list from cache %s: %v", cacheKey, err)
	}

	blogs, total, err := bu.blogRepository.ListBlogs(ctx, page, limit, field)
	if err != nil {
		return nil, 0, err
	}

	dataToCache := struct {
		Blogs []*domain.Blog `json:"blogs"`
		Total int64          `json:"total"`
	}{
		Blogs: blogs,
		Total: total,
	}
	blogListJSON, err := json.Marshal(dataToCache)
	if err == nil {
		bu.cacheUseCase.Set(ctx, cacheKey, blogListJSON, blogListCacheTTL)
	} else {
		log.Printf("Failed to marshal blog list for caching: %v", err)
	}

	return blogs, total, nil
}

func (bu *blogUsecase) SearchBlogs(ctx context.Context, query string) ([]*domain.Blog, error) {
	ctx, cancel := context.WithTimeout(ctx, bu.contextTimeout)
	defer cancel()

	cacheKey := fmt.Sprintf("blogs:search:%s", query)

	cachedBlogsBytes, err := bu.cacheUseCase.Get(ctx, cacheKey)
	if err == nil && cachedBlogsBytes != nil {
		var blogs []*domain.Blog
		if err := json.Unmarshal(cachedBlogsBytes, &blogs); err == nil {
			return blogs, nil
		}
		log.Printf("Failed to unmarshal cached search results %s: %v", cacheKey, err)
	} else if err != nil {
		log.Printf("Error getting search results from cache %s: %v", cacheKey, err)
	}

	blogs, err := bu.blogRepository.SearchBlogs(ctx, query)
	if err != nil {
		return nil, err
	}

	blogJSON, err := json.Marshal(blogs)
	if err == nil {
		bu.cacheUseCase.Set(ctx, cacheKey, blogJSON, blogListCacheTTL)
	} else {
		log.Printf("Failed to marshal search results for caching: %v", err)
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

	errChan := make(chan error, 1)
	var wg sync.WaitGroup

	if noReaction {
		err = bu.blogReactionRepository.AddReaction(ctx, reaction)
		if err != nil {
			log.Printf("Error adding reaction: %v", err)
			return err
		}

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

	go bu.cacheUseCase.Delete(context.Background(), fmt.Sprintf("blog:id:%s", blogID))

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
	go func() {
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
		} else {
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
	go bu.cacheUseCase.Delete(context.Background(), fmt.Sprintf("blog:id:%s", blogID))

	return nil
}

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
	go bu.cacheUseCase.Delete(context.Background(), fmt.Sprintf("comments:blog:%s", comment.BlogID))
	go bu.cacheUseCase.Delete(context.Background(), fmt.Sprintf("blog:id:%s", comment.BlogID))

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

	cacheKey := fmt.Sprintf("comments:blog:%s", blogID)

	cachedCommentsBytes, err := bu.cacheUseCase.Get(ctx, cacheKey)
	if err == nil && cachedCommentsBytes != nil {
		var comments []*domain.Comment
		if err := json.Unmarshal(cachedCommentsBytes, &comments); err == nil {
			return comments, nil
		}
		log.Printf("Failed to unmarshal cached comments %s: %v", cacheKey, err)
	} else if err != nil {
		log.Printf("Error getting comments from cache %s: %v", cacheKey, err)
	}

	comments, err := bu.blogCommentRepository.GetCommentsForBlog(ctx, blogID)
	if err != nil {
		return nil, err
	}

	commentsJSON, err := json.Marshal(comments)
	if err == nil {
		bu.cacheUseCase.Set(ctx, cacheKey, commentsJSON, commentListCacheTTL)
	} else {
		log.Printf("Failed to marshal comments for caching: %v", err)
	}

	return comments, nil
}

func (bu *blogUsecase) GetBlogsByUserID(ctx context.Context, userID string) ([]*domain.Blog, error) {
	ctx, cancel := context.WithTimeout(ctx, bu.contextTimeout)
	defer cancel()

	cacheKey := fmt.Sprintf("blogs:user:%s", userID)

	cachedBlogsBytes, err := bu.cacheUseCase.Get(ctx, cacheKey)
	if err == nil && cachedBlogsBytes != nil {
		var blogs []*domain.Blog
		if err := json.Unmarshal(cachedBlogsBytes, &blogs); err == nil {
			return blogs, nil
		}
		log.Printf("Failed to unmarshal cached user blogs %s: %v", cacheKey, err)
	} else if err != nil {
		log.Printf("Error getting user blogs from cache %s: %v", cacheKey, err)
	}

	blogs, err := bu.blogRepository.ListBlogsByAuthor(ctx, userID)
	if err != nil {
		return nil, err
	}

	blogsJSON, err := json.Marshal(blogs)
	if err == nil {
		bu.cacheUseCase.Set(ctx, cacheKey, blogsJSON, blogListCacheTTL)
	} else {
		log.Printf("Failed to marshal user blogs for caching: %v", err)
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

func (bu *blogUsecase) RemoveComment(ctx context.Context, commentID string) (error) {
	ctx, cancel := context.WithTimeout(ctx, bu.contextTimeout)
	defer cancel()
	
	err := bu.blogCommentRepository.DeleteComment(ctx, commentID)
	if err != nil {
		return err
	}
	go bu.cacheUseCase.InvalidatePrefix(context.Background(), "comments:blog:") 
	go bu.cacheUseCase.InvalidatePrefix(context.Background(), "blog:id:") 

	return nil
}
