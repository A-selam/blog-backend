package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI string
	DBName string
	GoogleClientID string
	GoogleClientSecret string
	JWTSecret string
	GeminiAPIKey string
}

func LoadConfig() (*Config, error){
	// Load .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: .env file not found or failed to load", err)
		return nil, err
	}
	
	// Retrieve environment variables
	MongoURI := os.Getenv("MONGODB_URI")
	if MongoURI == "" {
		log.Fatal("MONGODB_URI is not set")
		return nil, err
	}
	
	DBName := os.Getenv("DB_NAME")
	if DBName == "" {
		log.Fatal("DB_NAME is not set")
		return nil, err
	}
	
	GoogleClientID := os.Getenv("AUTH_CLIENT_ID")
	if GoogleClientID == "" {
		log.Fatal("GOOGLE_CLIENT_ID is not set")
		return nil, err
	}
	
	GoogleClientSecret := os.Getenv("AUTH_CLIENT_SECRET")
	if GoogleClientSecret == "" {
		log.Fatal("GOOGLE_CLIENT_SECRET is not set")
		return nil, err
	}
	
	JWTSecret := os.Getenv("JWT_SECRET")
	if JWTSecret == "" {
		log.Fatal("JWT_SECRET is not set")
		return nil, err
	}

	GeminiAPIKey := os.Getenv("GEMINI_API_KEY")
	if GeminiAPIKey == "" {
		log.Fatal("GEMINI_API_KEY is not set")
		return nil, err
	}

	return &Config{
		MongoURI : MongoURI,
		DBName : DBName,
		GoogleClientID : GoogleClientID,
		GoogleClientSecret : GoogleClientSecret,
		JWTSecret : JWTSecret,
		GeminiAPIKey : GeminiAPIKey,
	}, nil
}