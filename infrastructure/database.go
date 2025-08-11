package infrastructure

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func NewDatabase(mongoURI, dbName string) (*mongo.Client, *mongo.Database) {
	// Initialize MongoDB client
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().
		ApplyURI(mongoURI).
		SetMaxPoolSize(100).
		SetMinPoolSize(10).
		SetMaxConnIdleTime(30 * time.Second)

	client, err := mongo.Connect(clientOptions) 
	if err != nil {
		log.Fatal("MongoDB connection error:", err)
	}


	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}

	db := client.Database(dbName)

	if err := EnsureIndexes(ctx, db); err != nil {
		log.Fatalf("Failed to ensure MongoDB indexes: %v", err)
	}

	return client, db
}

// EnsureIndexes creates all necessary indexes for the collections.
// This function is idempotent and will only create indexes if they don't already exist.
func EnsureIndexes(ctx context.Context, db *mongo.Database) error {
	log.Println("Ensuring MongoDB indexes...")

	// --- Users Collection Indexes ---
	usersCollection := db.Collection("users")
	userIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "google_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetSparse(true), // Sparse for optional GoogleID
		},
		{
			Keys: bson.D{{Key: "created_at", Value: 1}}, // For sorting users
		},
	}
	if _, err := usersCollection.Indexes().CreateMany(ctx, userIndexes); err != nil {
		return fmt.Errorf("failed to create user indexes: %w", err)
	}
	log.Println("User indexes ensured.")

	// --- Blogs Collection Indexes ---
	blogsCollection := db.Collection("blogs")
	blogIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "author_id", Value: 1}}, // For fetching blogs by author
		},
		{
			Keys: bson.D{{Key: "tags", Value: 1}}, // Multi-key index for tags array
		},
		{
			Keys:    bson.D{{Key: "title", Value: "text"}}, // Text index for full-text search on title
			Options: options.Index().SetName("text_title"),
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}}, // For sorting by recent blogs
		},
		{
			Keys: bson.D{{Key: "view_count", Value: -1}}, // For sorting by most viewed
		},
		{
			Keys: bson.D{{Key: "like_count", Value: -1}}, // For sorting by most liked
		},
		{
			Keys: bson.D{{Key: "comment_count", Value: -1}}, // For sorting by most commented
		},
	}
	if _, err := blogsCollection.Indexes().CreateMany(ctx, blogIndexes); err != nil {
		return fmt.Errorf("failed to create blog indexes: %w", err)
	}
	log.Println("Blog indexes ensured.")

	// --- Comments Collection Indexes ---
	commentsCollection := db.Collection("comments")
	commentIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "blog_id", Value: 1}}, // For fetching comments for a specific blog
		},
		{
			Keys: bson.D{{Key: "author_id", Value: 1}}, // For checking comment author
		},
		{
			Keys: bson.D{{Key: "created_at", Value: 1}}, // For sorting comments
		},
	}
	if _, err := commentsCollection.Indexes().CreateMany(ctx, commentIndexes); err != nil {
		return fmt.Errorf("failed to create comment indexes: %w", err)
	}
	log.Println("Comment indexes ensured.")

	// --- Reactions Collection Indexes ---
	reactionsCollection := db.Collection("reactions")
	reactionIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "blog_id", Value: 1}, {Key: "user_id", Value: 1}},
			Options: options.Index().SetUnique(true), // Ensures one reaction per user per blog
		},
		{
			Keys: bson.D{{Key: "blog_id", Value: 1}}, // For fetching all reactions for a blog
		},
	}
	if _, err := reactionsCollection.Indexes().CreateMany(ctx, reactionIndexes); err != nil {
		return fmt.Errorf("failed to create reaction indexes: %w", err)
	}
	log.Println("Reaction indexes ensured.")

	// --- Refresh Tokens Collection Indexes ---
	refreshTokensCollection := db.Collection("refreshTokens")
	refreshTokenIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "token", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{
			Keys:    bson.D{{Key: "expires_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0), // TTL index: documents expire after 'expires_at' time
		},
	}
	if _, err := refreshTokensCollection.Indexes().CreateMany(ctx, refreshTokenIndexes); err != nil {
		return fmt.Errorf("failed to create refresh token indexes: %w", err)
	}
	log.Println("Refresh token indexes ensured.")

	// --- Password Reset Tokens Collection Indexes ---
	passwordResetTokensCollection := db.Collection("passwordResetTokens")
	passwordResetTokenIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "token", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "expires_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0), // TTL index
		},
		{
			Keys: bson.D{{Key: "used", Value: 1}}, // For filtering used tokens
		},
	}
	if _, err := passwordResetTokensCollection.Indexes().CreateMany(ctx, passwordResetTokenIndexes); err != nil {
		return fmt.Errorf("failed to create password reset token indexes: %w", err)
	}
	log.Println("Password reset token indexes ensured.")

	historyCollection := db.Collection("read_history") 
	historyIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "user_id", Value: 1}}, // If history is tied to users
		},
		{
			Keys: bson.D{{Key: "entity_id", Value: 1}}, // If history tracks actions on specific entities
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}}, // For fetching recent history
		},
	}
	if _, err := historyCollection.Indexes().CreateMany(ctx, historyIndexes); err != nil {
		return fmt.Errorf("failed to create history indexes: %w", err)
	}
	log.Println("History indexes ensured.")
	activationTokensCollection := db.Collection("RrefreshTokens") 
	activationTokenIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "token", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{
			Keys:    bson.D{{Key: "expires_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0), // TTL index
		},
	}
	if _, err := activationTokensCollection.Indexes().CreateMany(ctx, activationTokenIndexes); err != nil {
		return fmt.Errorf("failed to create refresh token indexes: %w", err)
	}
	log.Println("Refresh token indexes ensured.")

	return nil
}
