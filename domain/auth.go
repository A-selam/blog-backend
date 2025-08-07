package domain

import "context"

type IAuthUseCase interface {
	Register(ctx context.Context, user *User) (*User, error)
	Activate(ctx context.Context, tokenID string) error
	Login(ctx context.Context, email, password string) (*User, *TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
	RefreshToken(ctx context.Context, refreshToken string) (*User, *TokenPair, error)
	ForgotPassword(ctx context.Context, email string)  error
	ResetPassword(ctx context.Context, token, newPassword string) error
	FindOrCreateGoogleUser(ctx context.Context, email, username, profilePicture, googleID string) (*User, error)
	IssueTokenPair(ctx context.Context, user *User) (*TokenPair, error)
}

type IJWTService interface {
	GenerateToken(userID, username, email, role string) (string, error)
	ParseToken(tokenString string) (string, string, error)
}

type IPasswordService interface {
	HashPassword(password string) (string, error)
	ComparePassword(hashedPassword, plainPassword string) error 
}