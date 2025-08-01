package usecase

import (
	"blog-backend/domain"
	"context"
	"errors"
	"time"
	"github.com/google/uuid"
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

func (au *authUsecase) Register(ctx context.Context, username, email, password string) (*domain.User, error) {
	// TODO: implement this function
	return nil, nil
}

func (au *authUsecase) Login(ctx context.Context, email, password string) (*domain.User, *domain.TokenPair, error) {
	// TODO: implement this function
	return nil, nil, nil
}

func (au *authUsecase) Logout(ctx context.Context, refreshToken string) error {
	err := au.refreshTokenRepository.DeleteRefreshToken(ctx, refreshToken)
	if err != nil {
		return errors.New("failed to delete refresh token")
	}
	return nil
}

func (au *authUsecase) RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	// TODO: implement this function
	return nil, nil
}

func (au *authUsecase) ForgotPassword(ctx context.Context, email string) (token string, err error) {
	res, err := au.userRepository.GetUserByEmail(ctx, email)
	if err != nil{
		return "", domain.ErrInvalidUser
	}

	resetToken := &domain.PasswordResetToken{
		ID:        uuid.New().String(),
		UserID:    res.ID,
		Token:     uuid.New().String(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Used:      false,
		CreatedAt: time.Now(),
	}

	createdToken, err := au.resetTokenRepository.CreatePasswordResetToken(ctx,resetToken )
	if err != nil{
		return "", err
	}

	return createdToken.Token, nil
}

func (au *authUsecase) ResetPassword(ctx context.Context, token, newPassword string) error {	
	// TODO: implement this function
	resetToken, err := au.resetTokenRepository.GetPasswordResetToken(ctx, token)
	if err != nil {
		return domain.ErrInvalidPasswordResetToken
	}
	hashedPassword, err := au.passwordServices.HashPassword(newPassword)
	if err != nil {
		return err
	}
	updates := map[string]interface{}{
		"password_hash": hashedPassword,
		"updated_at": time.Now(),
	}
	err = au.userRepository.UpdateUser(ctx, resetToken.UserID, updates)
	if err != nil{
		return err
	}
	return nil
}
