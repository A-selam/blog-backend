package middleware

import (
	"blog-backend/domain"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func NewAuthMiddleware(jwtService domain.IJWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		userID, role, err := jwtService.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token", "details": err.Error()})
			c.Abort()
			return
		}
		// i saved the user id coz i need it for some bunch of me related api's
		c.Set("x-user-id", userID)
		c.Set("x-user-role", role)
		c.Next()
	}
}

func NewAdminMiddleware() gin.HandlerFunc{
	return func(c *gin.Context){
		role, exists := c.Get("x-user-role")
		if !exists || role != string(domain.Admin) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: Admins only"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func NewOptionalAuthMiddleware(jwtService domain.IJWTService) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
            tokenString := strings.TrimPrefix(authHeader, "Bearer ")
            userID, role, err := jwtService.ParseToken(tokenString)
            if err == nil {
                c.Set("x-user-id", userID)
                c.Set("x-user-role", role)
            }
        }
        c.Next()
    }
}
