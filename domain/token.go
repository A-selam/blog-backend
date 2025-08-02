package domain

import (
	"context"
	"errors"
	"time"
)

type RefreshToken struct {
	ID        string
	Token     string
	UserID    string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type TokenPair struct {
    AccessToken  string
    RefreshToken string
    ExpiresIn    time.Time 
}

type PasswordResetToken struct {
	ID        string
	UserID    string
	Token     string
	ExpiresAt time.Time
	Used      bool
	CreatedAt time.Time
}

type IRefreshTokenRepository interface {
	// Refresh Tokens
	CreateRefreshToken(ctx context.Context, userID, refToken string) (*RefreshToken, error)
	ReplaceRefreshToken(ctx context.Context, userID, refToken string) (*RefreshToken, error)
	GetRefreshToken(ctx context.Context, token string) (*RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, token string) error
	DeleteRefreshTokensForUser(ctx context.Context, userID string) error
}

type IResetTokenRepository interface {
	// Password Reset Tokens
	CreatePasswordResetToken(ctx context.Context, token *PasswordResetToken) (*PasswordResetToken, error)
	GetPasswordResetToken(ctx context.Context, token string) (*PasswordResetToken, error)
	MarkPasswordResetTokenUsed(ctx context.Context, token string) error
}

var (
	ErrTokenNotFound = errors.New("token not found")
	ErrInvalidBlogTitle = errors.New("invalid blog title")
	ErrInvalidBlogContent = errors.New("invalid blog content")
	ErrInvalidPasswordResetToken = errors.New("invalid password reset token")
	ErrPasswordResetTokenExpired = errors.New("password reset token expired")
	ErrTokenUsed = errors.New("token already used")
	ErrTokenExpired = errors.New("token expired")
	

)