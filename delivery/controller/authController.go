package controller

import (
	"blog-backend/domain"
	"github.com/gin-gonic/gin"
)

type AuthController struct {
	AuthUseCase domain.IAuthUseCase
}

func NewAuthController(au domain.IAuthUseCase) *AuthController {
	return &AuthController{
		AuthUseCase: au,
	}
}

func (ac *AuthController)Logout(c *gin.Context ) {
	var request struct {
		RefreshToken string `json:"refresh_token" binding:"required" `
	}
	if err:= c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request: refresh token is required"})	
	}

	_, exists := c.Get("userID")
	if !exists {
		c.JSON(400, gin.H{"error": "User ID not found in context"})
	}

	err := ac.AuthUseCase.Logout(c, request.RefreshToken)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to logout user."})
	}
	c.JSON(200, gin.H{"message": "User logged out successfully!"})
}

func (ac *AuthController)ForgotPassword(c *gin.Context) {
	var request struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request: email is required"})
		return
	}
	err := ac.AuthUseCase.ForgotPassword(c, request.Email)
	if err != nil {	
		c.JSON(500, gin.H{"error": "Failed to process forgot password request."})
		return	

	}
}