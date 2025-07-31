package middleware

import (
	"blog-backend/domain"

	"github.com/gin-gonic/gin"
)

func NewAuthMiddleware(jwtService domain.IJWTService) gin.HandlerFunc{
	return func (c *gin.Context){
		// TODO: implement the authentication middleware
	}
}

func NewAdminMiddleware() gin.HandlerFunc{
	return func(c *gin.Context){
		// TODO: implement the admin check middleware
	}
}
