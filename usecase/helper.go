package usecase
import (
	"blog-backend/domain"
	"context"
	"fmt"
)

func UpdateBlogMetricsAsync(ctx context.Context,brr domain.IReactionRepository,  br domain.IBlogRepository, reaction *domain.Reaction, blogID string, metric string, value int, results chan error)  {
	
	go func() {
		_, _, err := brr.CheckReactionExists(ctx, blogID, string(reaction.Type))
		if err != nil{
			results <- fmt.Errorf("failed to check if reaction exists:%w", err)
		}
	}()
	go func() {
		err := br.UpdateBlogMetrics(ctx, blogID, metric, value)
		if err != nil {
			results <- fmt.Errorf("failed to update blog metrics: %w", err)
			return
		}
		results <- nil
	}()
	go func() {
		err := brr.AddReaction(ctx, &domain.Reaction{
			BlogID: blogID,
			Type:   domain.ReactionType(metric),
	})
		if err != nil {
			results <- fmt.Errorf("failed to add reaction: %w", err)
			return
		}
		results <- nil
	}()

}

