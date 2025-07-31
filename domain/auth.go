package domain

import "context"

type IAuthUseCase interface {
	Register(ctx context.Context, username, email, password string) (*User, error)
	Login(ctx context.Context, email, password string) (*User, *TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
}

type IJWTService interface {
	GenerateToken(username, role string) (string, error)
	ParseToken(tokenString string) (string, string, error)
}

type IPasswordService interface {
	HashPassword(password string) (string, error)
	ComparePassword(hashedPassword, plainPassword string) error 
}