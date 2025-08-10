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
	"github.com/go-redis/redis/v8"

)

func main() {
	envConfig, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration:", err.Error())
	}

	googleConfig := config.GoogleConfig(envConfig.GoogleClientID, envConfig.GoogleClientSecret)

	client, db := infrastructure.NewDatabase(envConfig.MongoURI, envConfig.DBName)
	defer client.Disconnect(context.TODO())
	
	opts, err := redis.ParseURL(envConfig.RedisURL)
    if err != nil {
        log.Fatalf("Could not parse Redis URL: %v", err)
    }

    redisClient := redis.NewClient(opts)

    pong, err := redisClient.Ping(context.Background()).Result()
    if err != nil {
        log.Fatalf("Could not connect to Redis: %v", err)
    }
    log.Println("Connected to Redis:", pong)

	// Initialize services and repositories
    timeOut := 30 * time.Second
    jwtService := infrastructure.NewJWTService(envConfig.JWTSecret)
    passwordService := infrastructure.NewPasswordService()
	geminiService := infrastructure.NewGeminiService(envConfig.GeminiAPIKey)
	emailServices := infrastructure.NewEmailServices(envConfig.Email, envConfig.AppPassword)

	
	cacheRepo := repository.NewCacheRepository(redisClient)
	cacheUseCase := usecase.NewCacheUseCase(cacheRepo, redisClient, timeOut) 
	gu := usecase.NewGeminiUsecase(geminiService)
	gc := controller.NewGeminiController(gu)

	ur := repository.NewUserRepositoryFromDB(db)
	uu := usecase.NewUserUsecase(ur, timeOut, passwordService, cacheUseCase) 

	uc := controller.NewUserController(uu)
	bcr := repository.NewCommentRepositoryFromDB(db)
	brr := repository.NewReactionRepositoryFromDB(db)
	br := repository.NewBlogRepositoryFromDB(db)
<<<<<<< HEAD
	hr := repository.NewHistoryRepositoryFromDB(db)
	bu := usecase.NewBlogUsecase(br, brr, bcr,hr, geminiService,timeOut)
=======
	bu := usecase.NewBlogUsecase(br, brr, bcr, geminiService, timeOut, cacheUseCase) 
>>>>>>> df237cc94124ef968f73a256a16d06c209299f32
	bc := controller.NewBlogController(bu)
	resetTR := repository.NewResetTokenRepository(db)
	refreshTR := repository.NewRefreshTokenRepositoryFromDB(db)
	atr := repository.NewActivationTokenRepository(db)
	au := usecase.NewAuthUsecase(ur, refreshTR, resetTR, jwtService, passwordService, emailServices, atr, timeOut)
	ac := controller.NewAuthController(au, googleConfig)

    // Set up Gin router
    engine := gin.Default()
    route.Setup(ac, bc, uc, gc, jwtService, engine)

    // Start server
    if err := engine.Run("localhost:3000"); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}