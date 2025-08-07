package usecase

import (
	"blog-backend/domain"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type authUsecase struct {
	userRepository        domain.IUserRepository
	refreshTokenRepository domain.IRefreshTokenRepository
	resetTokenRepository domain.IResetTokenRepository
	jwtServices      domain.IJWTService
	passwordServices domain.IPasswordService
	contextTimeout        time.Duration
}

func NewAuthUsecase(
	userRepository        domain.IUserRepository,
	refreshTokenRepository domain.IRefreshTokenRepository,
	resetTokenRepository domain.IResetTokenRepository,
	jwtServices      domain.IJWTService,
	passwordServices domain.IPasswordService,
	timeout time.Duration,
) domain.IAuthUseCase {
	return &authUsecase{
		userRepository: userRepository,   
		refreshTokenRepository: refreshTokenRepository, 
		resetTokenRepository: resetTokenRepository,  
		jwtServices: jwtServices,
		passwordServices: passwordServices,
		contextTimeout: timeout, 
	}
}

func (au *authUsecase) Register(ctx context.Context, user *domain.User) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, au.contextTimeout)
	defer cancel()

	_, err := au.userRepository.GetUserByUsernameAndEmail(ctx, user.Username, user.Email)
	if err == nil {
		return nil, fmt.Errorf("username is already taken")
	}
	if err != mongo.ErrNoDocuments {
		return nil, err 
	}

	hashedPassword, err := au.passwordServices.HashPassword(user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}
	user.PasswordHash = hashedPassword

	createdUser, err := au.userRepository.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}
	
	return createdUser, nil
}

func (au *authUsecase) Login(ctx context.Context, email, password string) (*domain.User, *domain.TokenPair, error) {
	ctx, cancel := context.WithTimeout(ctx, au.contextTimeout)
	defer cancel()

	user, err := au.userRepository.GetUserByEmail(ctx, email)
	if err == mongo.ErrNoDocuments || au.passwordServices.ComparePassword(user.PasswordHash, password) != nil{
		fmt.Println(user)
		return nil, nil, fmt.Errorf("Invalid email or password")
	}
	if err != nil {
		fmt.Println(err.Error())
		return nil, nil, err
	}

	jwtToken, err := au.jwtServices.GenerateToken(user.ID, user.Username, user.Email, string(user.Role))
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to generate JWT token: %v", err)
	}

	refToken, err := generateRefreshToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to generate refresh token: %v", err)
	}

	refreshToken, err := au.refreshTokenRepository.CreateRefreshToken(ctx, user.ID, refToken,)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to create refresh token: %v", err)
	}

	tokenPair := &domain.TokenPair{
		AccessToken: jwtToken,
		RefreshToken: refreshToken.Token,
		ExpiresIn: refreshToken.ExpiresAt,
	}
	
	return user, tokenPair, nil
}

func (au *authUsecase) Logout(ctx context.Context, refreshToken string) error {
	err := au.refreshTokenRepository.DeleteRefreshToken(ctx, refreshToken)
	if err != nil {
		return errors.New("failed to delete refresh token")
	}
	return nil
}

func (au *authUsecase) RefreshToken(ctx context.Context, refreshToken string) (*domain.User, *domain.TokenPair, error) {
	ctx, cancel := context.WithTimeout(ctx, au.contextTimeout)
	defer cancel()

	refreshTokenData, err := au.refreshTokenRepository.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get refresh token: %v", err)
	}
	
	if refreshTokenData == nil {
	return nil, nil, errors.New("refresh token not found")
}

	if refreshTokenData.ExpiresAt.Before(time.Now()) {
		if err := au.Logout(ctx, refreshTokenData.Token); err != nil {
			log.Printf("Failed to delete expired refresh token: %v", err)
		}
	return nil, nil, errors.New("refresh token has expired")
}

	
	user, err := au.userRepository.GetUserByID(ctx, refreshTokenData.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user: %v", err)
	}	

	jwtToken, err := au.jwtServices.GenerateToken(user.ID, user.Username, user.Email, string(user.Role))
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to generate JWT token: %v", err)
	}

	newRefreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to generate new refresh token: %v", err)
	}

	refToken, err := au.refreshTokenRepository.ReplaceRefreshToken(ctx, user.ID, newRefreshToken)

	tokenPair := &domain.TokenPair{
		AccessToken: jwtToken,
		RefreshToken: refToken.Token,
		ExpiresIn: refToken.ExpiresAt,
	}
	
	return user, tokenPair, nil
}

func (au *authUsecase) FindOrCreateGoogleUser(ctx context.Context, email, username, profilePicture, googleID string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, au.contextTimeout)
	defer cancel()

	found := true

	user, err := au.userRepository.GetUserByGoogleID(ctx, googleID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			found = false
		} else {
			return nil, err
		}
	}

	if !found {
		user = &domain.User{
			GoogleID: googleID,
			Email: email,
			Username: username,
			ProfilePicture: profilePicture,
			Role: domain.RegularUser,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		user, err = au.userRepository.CreateUser(ctx, user)
		if err != nil {
			return nil, err
		}
	}
	return user, nil
}

func (au *authUsecase) IssueTokenPair(ctx context.Context, user *domain.User) (*domain.TokenPair, error) {
	jwtToken, err := au.jwtServices.GenerateToken(user.ID, user.Username, user.Email, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("Failed to generate JWT token: %v", err)
	}

	refToken, err := generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %v", err)
	}

	refreshToken, err := au.refreshTokenRepository.CreateRefreshToken(ctx, user.ID, refToken,)
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token  %v", err)
	}

	tokenPair := &domain.TokenPair{
		AccessToken: jwtToken,
		RefreshToken: refreshToken.Token,
		ExpiresIn: refreshToken.ExpiresAt,
	}

	return tokenPair, nil
}

func (au *authUsecase) ForgotPassword(ctx context.Context, email string) (token string, err error) {
	res, err := au.userRepository.GetUserByEmail(ctx, email)
	if err != nil{
		return "", domain.ErrInvalidUser
	}

	resetToken := &domain.PasswordResetToken{
		ID:			uuid.New().String(),
		UserID:    res.ID,
		Token:     uuid.New().String(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Used:      false,
		CreatedAt: time.Now(),
	}

	createdToken, err := au.resetTokenRepository.CreatePasswordResetToken(ctx,resetToken)
	if err != nil{
		return "", err
	}

	return createdToken.Token, nil
}


func (au *authUsecase) ResetPassword(ctx context.Context, token, newPassword string) error {
    ctx, cancel := context.WithTimeout(ctx, au.contextTimeout)
    defer cancel()

    resetToken, err := au.resetTokenRepository.GetPasswordResetToken(ctx, token)
    if err != nil {
        log.Printf("Error fetching reset token: %v", err)
        return err
    }

    hashedPassword, err := au.passwordServices.HashPassword(newPassword)
    if err != nil {
        log.Printf("Error hashing password: %v", err)
        return fmt.Errorf("failed to hash password: %v", err)
    }

    updates := map[string]interface{}{
        "password_hash": hashedPassword,
        "updated_at":    time.Now(),
    }
    err = au.userRepository.UpdateUser(ctx, resetToken.UserID, updates)
	log.Print(resetToken.UserID)
    if err != nil {
        log.Printf("Error updating user: %v", err)
        return fmt.Errorf("failed to update user: %v", err)
    }

    err = au.resetTokenRepository.MarkPasswordResetTokenUsed(ctx, token)
    if err != nil {
        log.Printf("Error marking token as used: %v", err)
        return fmt.Errorf("failed to mark reset token as used: %v", err)
    }

    return nil
}

func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}