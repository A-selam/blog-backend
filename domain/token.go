package domain

import (
	"context"
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
    ExpiresIn    int64 
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
	CreateRefreshToken(ctx context.Context, token *RefreshToken) (*RefreshToken, error)
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