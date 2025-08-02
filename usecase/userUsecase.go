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
	user, err := uu.userRepository.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (uu *userUsecase) UpdateProfile(ctx context.Context, userID string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()

	err := uu.userRepository.UpdateUser(ctx, userID, updates)
	if err != nil {
		return err
	}

	return nil
}

// Admin Only
func (uu *userUsecase) PromoteToAdmin(ctx context.Context, targetUserID string) error {
	ctx, cancel := context.WithTimeout(ctx, uu.contextTimeout)
	defer cancel()

	err := uu.userRepository.PromoteToAdmin(ctx, targetUserID)
	if err != nil {
		return err
	}

	return nil
}

func (uu *userUsecase) DemoteToUser(ctx context.Context, targetUserID string) error {
	ctx, cancel := context.WithTimeout(ctx, uu.contextTimeout)
	defer cancel()

	err := uu.userRepository.DemoteToUser(ctx, targetUserID)
	if err != nil {
		return err
	}

	return nil
}
