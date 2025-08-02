package controller

import (
	"blog-backend/domain"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	UserUseCase domain.IUserUseCase
}

func NewUserController(uu domain.IUserUseCase) *UserController{
	return &UserController{
		UserUseCase: uu,
	}
}

func (uc *UserController) GetUserByID(c *gin.Context) {
	userID := c.Param("id")

	user, err := uc.UserUseCase.GetProfile(c, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (uc *UserController) GetCurrentUserProfile(c *gin.Context) {
	userID, ok := c.Get("x-user-id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context."})
		return
	}

	user, err := uc.UserUseCase.GetProfile(c, userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (uc *UserController) UpdateCurrentUserProfile(c *gin.Context) {
	userID, ok := c.Get("x-user-id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context."})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body."})
		return
	}

	err := uc.UserUseCase.UpdateProfile(c, userID.(string), updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user profile."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User profile updated successfully."})
}

func (uc *UserController) PromoteUser(c *gin.Context){
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required."})
		return
	}

	err := uc.UserUseCase.PromoteToAdmin(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to promote user."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User promoted successfully."})
}

func (uc *UserController) DemoteUser(c *gin.Context){
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required."})
		return
	}

	err := uc.UserUseCase.DemoteToUser(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to demote user."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User demoted successfully."})
}