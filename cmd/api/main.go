package main

import (
	"blog-backend/delivery/controller"
	"blog-backend/delivery/route"
	"blog-backend/infrastructure"
	"blog-backend/repository"
	"blog-backend/usecase"
	"context"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Println("Warning: .env file not found or failed to load", err)
	}

	// Retrieve environment variables
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("MONGODB_URI is not set")
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		log.Fatal("DB_NAME is not set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is not set")
	}

	client, db := infrastructure.NewDatabase(mongoURI, dbName)
	defer client.Disconnect(context.TODO())
	// Initialize services and repositories
    timeOut := 30 * time.Second
    jwtService := infrastructure.NewJWTService(jwtSecret)
    passwordService := infrastructure.NewPasswordService()

	ur := repository.NewUserRepositoryFromDB(db)
	uu := usecase.NewUserUsecase(ur, timeOut,passwordService)
	uc := controller.NewUserController(uu)
	bmr := repository.NewBlogMetricsRepositoryFromDB(db)
	bcr := repository.NewCommentRepositoryFromDB(db)
	brr := repository.NewReactionRepositoryFromDB(db)
	br := repository.NewBlogRepositoryFromDB(db)
	bu := usecase.NewBlogUsecase(br, brr, bcr, timeOut, bmr)
	bc := controller.NewBlogController(bu)

	resetTR := repository.NewResetTokenRepository(db)
	refreshTR := repository.NewRefreshTokenRepositoryFromDB(db)
	au := usecase.NewAuthUsecase(ur, refreshTR, resetTR, jwtService, passwordService, timeOut)
	ac := controller.NewAuthController(au)

    // Set up Gin router
    engine := gin.Default()
    route.Setup(ac, bc, uc, jwtService, engine)

    // Start server
    if err := engine.Run("localhost:3000"); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}