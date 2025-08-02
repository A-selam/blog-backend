package route

import (
	"blog-backend/delivery/controller"
	"blog-backend/domain"
	"blog-backend/infrastructure/middleware"

	"github.com/gin-gonic/gin"
)

func Setup(ac *controller.AuthController, bc *controller.BlogController, uc *controller.UserController, jwtService domain.IJWTService, engine *gin.Engine) {
	// ============ Public Routes ============
	publicRouter := engine.Group("/api")
	NewAuthRouter(ac, publicRouter)
	NewBlogRouter(bc, publicRouter) // Public blog routes (read-only)

	// ============ Protected Routes (User) ============
	userRouter := engine.Group("/api")
	userRouter.Use(middleware.NewAuthMiddleware(jwtService))
	NewUserRouter(uc, userRouter)
	NewBlogAuthRouter(bc, userRouter) // Authenticated blog routes (create/update/delete)

	// ============ Admin Routes ============
	adminRouter := engine.Group("/api/admin")
	adminRouter.Use(middleware.NewAuthMiddleware(jwtService))
	adminRouter.Use(middleware.NewAdminMiddleware())
	NewAdminRouter(uc, bc, adminRouter)
}

func NewAuthRouter(handler *controller.AuthController, group *gin.RouterGroup) {
	group.POST("/auth/register", handler.Register)
	group.POST("/auth/login", handler.Login)
	group.POST("/auth/forgot-password", )
	group.POST("/auth/reset-password", )
	group.POST("/auth/refresh", handler.RefreshToken)
}

func NewUserRouter(handler *controller.UserController, group *gin.RouterGroup) {
	group.GET("/users/me", handler.GetCurrentUserProfile)
	group.PATCH("/users/me",handler.UpdateCurrentUserProfile )
	group.GET("/users/:id", handler.GetUserByID)
}

func NewBlogRouter(handler *controller.BlogController, group *gin.RouterGroup) {
	group.GET("/blogs", )
	group.GET("/blogs/:id", )
	group.GET("/blogs/search", )
	group.GET("/blogs/:id/comments", )
	group.GET("/blogs/:id/metrics", )
}

func NewBlogAuthRouter(handler *controller.BlogController, group *gin.RouterGroup) {
	group.POST("/blogs", handler.CreateBlog)
	group.PATCH("/blogs/:id", )
	group.DELETE("/blogs/:id", )
	group.POST("/blogs/:id/like", )
	group.POST("/blogs/:id/dislike", )
	group.DELETE("/blogs/:id/reaction", )
	group.POST("/blogs/:id/comments", )
}

func NewAdminRouter(userHandler *controller.UserController, blogHandler *controller.BlogController, group *gin.RouterGroup) {
	// User Management
	group.GET("/users",)
	group.POST("/users/:id/promote", userHandler.PromoteUser)
	group.POST("/users/:id/demote", userHandler.DemoteUser)
	group.DELETE("/users/:id", )

	// Blog Moderation
	group.DELETE("/blogs/:id", )
}