package usecase

import (
	"blog-backend/domain"
	"context"
	"encoding/json" 
	"fmt"
	"log"
	"time"
)

const (
	userProfileCacheTTL = 15 * time.Minute // TTL for user profiles
	userListCacheTTL    = 5 * time.Minute // TTL for lists of users
)

type userUsecase struct {
	userRepository   domain.IUserRepository
	contextTimeout   time.Duration
	passwordServices domain.IPasswordService
	cacheUseCase     domain.ICacheUseCase 
}

func NewUserUsecase(
	userRepository domain.IUserRepository,
	timeout time.Duration,
	passwordServices domain.IPasswordService,
	cacheUseCase domain.ICacheUseCase, 
) domain.IUserUseCase {
	return &userUsecase{
		userRepository:   userRepository,
		contextTimeout:   timeout,
		passwordServices: passwordServices,
		cacheUseCase:     cacheUseCase, 
	}
}

func (uu *userUsecase) GetProfile(ctx context.Context, userID string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, uu.contextTimeout)
	defer cancel()

	cacheKey := fmt.Sprintf("user:id:%s", userID)

	cachedUserBytes, err := uu.cacheUseCase.Get(ctx, cacheKey)
	if err == nil && cachedUserBytes != nil {
		var user domain.User
		if err := json.Unmarshal(cachedUserBytes, &user); err == nil {
			return &user, nil 
		}
		log.Printf("Failed to unmarshal cached user %s: %v", userID, err)
	} else if err != nil {
		log.Printf("Error getting user from cache %s: %v", userID, err)
	}

	user, err := uu.userRepository.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	userJSON, err := json.Marshal(user)
	if err == nil {
		uu.cacheUseCase.Set(ctx, cacheKey, userJSON, userProfileCacheTTL)
	} else {
		log.Printf("Failed to marshal user %s for caching: %v", userID, err)
	}

	return user, nil
}

func (uu *userUsecase) UpdateProfile(ctx context.Context, userID string, updates map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, uu.contextTimeout)
	defer cancel()
	updates["updated_at"] = time.Now()

	if newPass, ok := updates["password"].(string); ok {
		hashedPass, err := uu.passwordServices.HashPassword(newPass)
		if err != nil {
			return err
		}
		updates["password_hash"] = hashedPass
		delete(updates, "password")
	}

	err := uu.userRepository.UpdateUser(ctx, userID, updates)
	if err != nil {
		return err
	}

	go uu.cacheUseCase.Delete(context.Background(), fmt.Sprintf("user:id:%s", userID))
	go uu.cacheUseCase.InvalidatePrefix(context.Background(), "users:list:")

	return nil
}

func (uu *userUsecase) PromoteToAdmin(ctx context.Context, targetUserID string) error {
	ctx, cancel := context.WithTimeout(ctx, uu.contextTimeout)
	defer cancel()

	err := uu.userRepository.PromoteToAdmin(ctx, targetUserID)
	if err != nil {
		return err
	}

	go uu.cacheUseCase.Delete(context.Background(), fmt.Sprintf("user:id:%s", targetUserID))
	go uu.cacheUseCase.InvalidatePrefix(context.Background(), "users:list:") 
	return nil
}

func (uu *userUsecase) DemoteToUser(ctx context.Context, targetUserID string) error {
	ctx, cancel := context.WithTimeout(ctx, uu.contextTimeout)
	defer cancel()

	err := uu.userRepository.DemoteToUser(ctx, targetUserID)
	if err != nil {
		return err
	}

	go uu.cacheUseCase.Delete(context.Background(), fmt.Sprintf("user:id:%s", targetUserID))
	go uu.cacheUseCase.InvalidatePrefix(context.Background(), "users:list:") 

	return nil
}

func (uu *userUsecase) GetUsers(ctx context.Context, page, limit int) ([]*domain.User, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, uu.contextTimeout)
	defer cancel()

	cacheKey := fmt.Sprintf("users:list:page:%d:limit:%d", page, limit)

	cachedUsersBytes, err := uu.cacheUseCase.Get(ctx, cacheKey)
	if err == nil && cachedUsersBytes != nil {
		var cachedData struct {
			Users []*domain.User `json:"users"`
			Total int64          `json:"total"`
		}
		if err := json.Unmarshal(cachedUsersBytes, &cachedData); err == nil {
			return cachedData.Users, cachedData.Total, nil
		}
		log.Printf("Failed to unmarshal cached user list %s: %v", cacheKey, err)
	} else if err != nil {
		log.Printf("Error getting user list from cache %s: %v", cacheKey, err)
	}

	users, total, err := uu.userRepository.GetUsers(ctx, page, limit)
	if err != nil {
		return nil, 0, err
	}

	dataToCache := struct {
		Users []*domain.User `json:"users"`
		Total int64          `json:"total"`
	}{
		Users: users,
		Total: total,
	}
	userListJSON, err := json.Marshal(dataToCache)
	if err == nil {
		uu.cacheUseCase.Set(ctx, cacheKey, userListJSON, userListCacheTTL)
	} else {
		log.Printf("Failed to marshal user list for caching: %v", err)
	}

	return users, total, nil
}

func (uu *userUsecase) DeleteUser(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, uu.contextTimeout)
	defer cancel()

	err := uu.userRepository.DeleteUser(ctx, id)
	if err != nil {
		return err
	}

	go uu.cacheUseCase.Delete(context.Background(), fmt.Sprintf("user:id:%s", id))
	go uu.cacheUseCase.InvalidatePrefix(context.Background(), "users:list:")

	return nil
}