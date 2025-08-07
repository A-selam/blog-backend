package main

import (
	"blog-backend/config"
	"blog-backend/delivery/controller"
	"blog-backend/delivery/route"
	"blog-backend/infrastructure"
	"blog-backend/repository"
	"blog-backend/usecase"
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	envConfig, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration:", err.Error())
	}

	googleConfig := config.GoogleConfig(envConfig.GoogleClientID, envConfig.GoogleClientSecret)

	client, db := infrastructure.NewDatabase(envConfig.MongoURI, envConfig.DBName)
	defer client.Disconnect(context.TODO())

	// Initialize services and repositories
    timeOut := 30 * time.Second
    jwtService := infrastructure.NewJWTService(envConfig.JWTSecret)
    passwordService := infrastructure.NewPasswordService()
	geminiService := infrastructure.NewGeminiService(envConfig.GeminiAPIKey)

	gu := usecase.NewGeminiUsecase(geminiService)
	gc := controller.NewGeminiController(gu)

	ur := repository.NewUserRepositoryFromDB(db)
	uu := usecase.NewUserUsecase(ur, timeOut,passwordService)
	uc := controller.NewUserController(uu)
	bcr := repository.NewCommentRepositoryFromDB(db)
	brr := repository.NewReactionRepositoryFromDB(db)
	br := repository.NewBlogRepositoryFromDB(db)
	bu := usecase.NewBlogUsecase(br, brr, bcr, geminiService,timeOut)
	bc := controller.NewBlogController(bu)
	resetTR := repository.NewResetTokenRepository(db)
	refreshTR := repository.NewRefreshTokenRepositoryFromDB(db)
	au := usecase.NewAuthUsecase(ur, refreshTR, resetTR, jwtService, passwordService, timeOut)
	ac := controller.NewAuthController(au, googleConfig)

    // Set up Gin router
    engine := gin.Default()
    route.Setup(ac, bc, uc, gc, jwtService, engine)

    // Start server
    if err := engine.Run("localhost:3000"); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}