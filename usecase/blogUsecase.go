package usecase

import (
	"blog-backend/domain"
	"context"
	"errors"
	"fmt"
	"time"
)

type blogUsecase struct {
	blogRepository         domain.IBlogRepository
	blogReactionRepository domain.IReactionRepository
	blogCommentRepository  domain.ICommentRepository
	geminiServices         domain.IGeminiService
	contextTimeout         time.Duration
}

// RemoveComment implements domain.IBlogUseCase.
func (bu *blogUsecase) RemoveComment(ctx context.Context, commentID string) error {
	panic("unimplemented")
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
	err = bu.blogRepository.UpdateBlogMetrics(ctx, blogID, "view_count", 1)
	if err != nil {
		return nil, err
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

// Reactions
func (bu *blogUsecase) AddReaction(ctx context.Context, blogID, userID string, reactionType string) error {
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
		return err
	}

	if noReaction {
		err = bu.blogReactionRepository.AddReaction(ctx, reaction)
		if err != nil {
			return err
		}
		if reactionType == string(domain.Like) {
			err = bu.blogRepository.UpdateBlogMetrics(ctx, blogID, string(domain.LikeCountField), 1)
			if err != nil {
				return err
			}
		} else {
			err = bu.blogRepository.UpdateBlogMetrics(ctx, blogID, string(domain.DislikeCountField), 1)
			if err != nil {
				return err
			}
		}
		return nil
	} else if string(rxn.Type) != reactionType {
		err = bu.blogReactionRepository.UpdateReaction(ctx, blogID, userID, domain.ReactionType(reactionType))
		if err != nil {
			return err
		}

		if reactionType == string(domain.Like) {
			err = bu.blogRepository.UpdateBlogMetrics(ctx, blogID, string(domain.LikeCountField), 1)
			if err != nil {
				return err
			}
			err = bu.blogRepository.UpdateBlogMetrics(ctx, blogID, string(domain.DislikeCountField), -1)
			if err != nil {
				return err
			}
		} else if reactionType == string(domain.Dislike) {
			err = bu.blogRepository.UpdateBlogMetrics(ctx, blogID, string(domain.DislikeCountField), 1)
			if err != nil {
				return err
			}
			err = bu.blogRepository.UpdateBlogMetrics(ctx, blogID, string(domain.LikeCountField), -1)
			if err != nil {
				return err
			}
		}
		return nil
	}

	return errors.New("reaction already exists with the same type")
}

func (bu *blogUsecase) RemoveReaction(ctx context.Context, blogID, userID string) error {
	ctx, cancel := context.WithTimeout(ctx, bu.contextTimeout)
	defer cancel()
	rxn, noReaction, err := bu.blogReactionRepository.CheckReactionExists(ctx, blogID, userID)
	if err != nil {
		return err
	}
	if noReaction {
		return errors.New("no reaction found to remove")
	}
	fmt.Println("Reaction found:", rxn.Type, domain.Like)
	if string(rxn.Type) == string(domain.Like) {
		err = bu.blogRepository.UpdateBlogMetrics(ctx, blogID, string(domain.LikeCountField), -1)
		if err != nil {
			return err
		}
	} else if rxn.Type == domain.Dislike {
		err = bu.blogRepository.UpdateBlogMetrics(ctx, blogID, string(domain.DislikeCountField), -1)
		if err != nil {
			return err
		}
	} else {
		return errors.New("reaction type not recognized")
	}

	err = bu.blogReactionRepository.RemoveReaction(ctx, blogID, userID)
	if err != nil {
		return err
	}

	return nil
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
	if err != nil {
		return nil, err
	}
	err = bu.blogRepository.UpdateBlogMetrics(ctx, blogID, "comment_count", 1)
	if err != nil {
		return nil, err
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
