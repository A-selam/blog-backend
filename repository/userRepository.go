package repository

import (
	"blog-backend/domain"
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type userRepository struct {
	database   *mongo.Database
	collection string
}

func NewUserRepositoryFromDB(db *mongo.Database) domain.IUserRepository {
	return &userRepository{
		database:   db,
		collection: "users",
	}
}

func (ur userRepository) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	collection := ur.database.Collection(ur.collection)

	userDTO := DomainToDTO(*user)
	insertedResult, err := collection.InsertOne(ctx, userDTO)
	if err != nil {
		return nil, err
	}
	userDTO.ID = insertedResult.InsertedID.(bson.ObjectID)

	return DTOToDomain(userDTO), nil
}

func (ur userRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	collection := ur.database.Collection(ur.collection)

	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %v", err)
	}

	filter := bson.D{{Key: "_id", Value: oid}}

	var user *UserDTO
	err = collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, err
	}
	return DTOToDomain(user), nil
}

func (ur userRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	collection := ur.database.Collection(ur.collection)
	filter := bson.D{{Key: "email", Value: email}}

	var user *UserDTO
	err := collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, err
	}
	
	return DTOToDomain(user), nil
}

func (ur userRepository) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	// TODO: Implement the function
	return nil, nil
}

func (ur userRepository) GetUserByUsernameAndEmail(ctx context.Context, username, email string) (*domain.User, error){
	collection := ur.database.Collection(ur.collection)
	filter := bson.D{{Key: "username", Value:username}, {Key: "email", Value: email}}

	var user *UserDTO
	err := collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, err
	}
	return DTOToDomain(user), nil
}


func (ur userRepository) UpdateUser(ctx context.Context, id string, updates map[string]interface{}) error {

    if id == "" {
        return domain.ErrInvalidUserID
    }
    oid, err := bson.ObjectIDFromHex(id)
    if err != nil {
        log.Printf("Invalid ObjectID: %v", err)
        return fmt.Errorf("invalid user ID: %v", err)
    }

    updates["updated_at"] = time.Now()

    filter := bson.M{"_id": oid}
    update := bson.M{"$set": updates}

    collection := ur.database.Collection(ur.collection)
    result, err := collection.UpdateOne(ctx, filter, update)
    if err != nil {
        log.Printf("Update error: %v (collection: %s)", err, ur.collection)
        return fmt.Errorf("failed to update user: %v", err)
    }

    if result.MatchedCount == 0 {
        // var doc bson.M
        // err := collection.FindOne(ctx, filter).Decode(&doc)
        // if err == mongo.ErrNoDocuments {
        //     log.Printf("Document with _id=%v not found in collection '%s'", oid, ur.collection)
        //     return domain.ErrUserNotFound
        // } else if err != nil {
        //     log.Printf("Find error: %v", err)
        //     return fmt.Errorf("failed to check document existence: %v", err)
        // }
        // log.Printf("Document exists but wasn't updated (possible no-op)")
        return domain.ErrUserNotFound 
    }

    return nil
}

func (ur userRepository) DeleteUser(ctx context.Context, id string) error {
	// TODO: Implement the function
	return nil
}

// Profile Management 
func (ur userRepository) UpdateProfile(ctx context.Context, userID string, bio, profilePicture, contactInfo string) error {
	// TODO: Implement the function
	return nil
}

// Admin Actions 
func (ur userRepository) PromoteToAdmin(ctx context.Context, userID string) error {
	collection := ur.database.Collection(ur.collection)

	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %v", err)
	}

	filter := bson.D{{Key: "_id", Value: oid}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "role", Value: string(domain.Admin)}}}}

	updateResult, err := collection.UpdateOne(ctx, filter, update)
	if err != nil || updateResult.MatchedCount == 0 { 
		return fmt.Errorf("failed to promote user to admin: %v", err)
	}

	return nil
}

func (ur userRepository) DemoteToUser(ctx context.Context, userID string) error {
	collection := ur.database.Collection(ur.collection)

	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %v", err)
	}

	filter := bson.D{{Key: "_id", Value: oid}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "role", Value: string(domain.RegularUser)}}}}

	updateResult, err := collection.UpdateOne(ctx, filter, update)
	if err != nil || updateResult.MatchedCount == 0 { 
		return fmt.Errorf("failed to demote user to admin: %v", err)
	}

	return nil
}

// DTOs

type UserDTO struct {
    ID             bson.ObjectID `bson:"_id,omitempty"`
    Username       string             `bson:"username"` // unique
    Email          string             `bson:"email"`    // unique
    PasswordHash   string             `bson:"password_hash"`
    Role           string             `bson:"role"`
    CreatedAt      time.Time          `bson:"created_at"`
    UpdatedAt      time.Time          `bson:"updated_at"`
    Bio            string             `bson:"bio,omitempty"`
    ProfilePicture string             `bson:"profile_picture,omitempty"`
    ContactInfo    string             `bson:"contact_info,omitempty"`
}

// DTO mapper

func DTOToDomain(d *UserDTO) *domain.User {
	return &domain.User{
		ID:             (d.ID).Hex(),
		Username:       d.Username,
		Email:          d.Email,
		PasswordHash:   d.PasswordHash,
		Role:           domain.Role(d.Role),
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
		Bio:            d.Bio,
		ProfilePicture: d.ProfilePicture,
		ContactInfo:    d.ContactInfo,
	}
}

func DomainToDTO(u domain.User) *UserDTO{
	return &UserDTO{
		Username: u.Username,
		Email: u.Email,
		PasswordHash: u.PasswordHash,
		Role: string(u.Role),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Bio: u.Bio,
		ProfilePicture: u.ProfilePicture,
		ContactInfo: u.ContactInfo,
	}
}