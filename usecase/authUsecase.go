package usecase

import (
	"blog-backend/domain"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

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
		return nil, fmt.Errorf("Username is already taken.")
	}
	if err != mongo.ErrNoDocuments {
		return nil, err 
	}

	hashedPassword, err := au.passwordServices.HashPassword(user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("Failed to hash password: %v", err)
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
	// TODO: implement this function
	return nil
}

func (au *authUsecase) RefreshToken(ctx context.Context, refreshToken string) (*domain.User, *domain.TokenPair, error) {
	ctx, cancel := context.WithTimeout(ctx, au.contextTimeout)
	defer cancel()

	refreshTokenData, err := au.refreshTokenRepository.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to get refresh token: %v", err)
	}
	
	if refreshTokenData == nil || refreshTokenData.ExpiresAt.Before(time.Now()) {
		return nil, nil, fmt.Errorf("Invalid or expired refresh token")	
	}
	
	user, err := au.userRepository.GetUserByID(ctx, refreshTokenData.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to get user: %v", err)
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

func (au *authUsecase) ForgotPassword(ctx context.Context, email string) error {
	// TODO: implement this function
	return nil
}

func (au *authUsecase) ResetPassword(ctx context.Context, token, newPassword string) error {
	// TODO: implement this function
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