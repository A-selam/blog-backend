package domain

import (
	"context"
	"time"
)

type Role string

const (
	RegularUser Role = "User"
	Admin       Role = "Admin"
)

type User struct {
	ID           string 
	Username     string 
	Email        string 
	PasswordHash string 
	Role         Role   
	CreatedAt    time.Time 
	UpdatedAt    time.Time 

	// Profile Info
	Bio            string
	ProfilePicture string 
	ContactInfo    string
}

type Login struct {
	Email string
	Password string
}

type IUserRepository interface {
	// User Management
	CreateUser(ctx context.Context, user *User) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	GetUserByUsernameAndEmail(ctx context.Context, username, email string) (*User, error)
	UpdateUser(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteUser(ctx context.Context, id string) error

	// Profile Management
	UpdateProfile(ctx context.Context, userID string, bio, profilePicture, contactInfo string) error

	// Admin Actions
	PromoteToAdmin(ctx context.Context, userID string) error
	DemoteToUser(ctx context.Context, userID string) error
}

type IUserUseCase interface {
	GetProfile(ctx context.Context, userID string) (*User, error)
	UpdateProfile(ctx context.Context, userID string, updates map[string]interface{}) error

	// Admin Only
	PromoteToAdmin(ctx context.Context, adminID, targetUserID string) error
	DemoteToUser(ctx context.Context, adminID, targetUserID string) error
}