package usecase

import (
	"blog-backend/domain"
	"context"
	"time"
)

type userUsecase struct {
	userRepository   domain.IUserRepository
	contextTimeout   time.Duration
}

func NewUserUsecase(
	userRepository   domain.IUserRepository,
	timeout time.Duration,
) domain.IUserUseCase {
	return &userUsecase{
		userRepository:   userRepository,
		contextTimeout:   timeout,
	}
}

func (uu *userUsecase) GetProfile(ctx context.Context, userID string) (*domain.User, error) {
	// TODO: implement this function
	return nil, nil
}

func (uu *userUsecase) UpdateProfile(ctx context.Context, userID string, updates map[string]interface{}) error {
	// TODO: implement this function
	return nil
}

// Admin Only
func (uu *userUsecase) PromoteToAdmin(ctx context.Context, adminID, targetUserID string) error {
	// TODO: implement this function
	return nil
}

func (uu *userUsecase) DemoteToUser(ctx context.Context, adminID, targetUserID string) error {
	// TODO: implement this function
	return nil
}
