package controller

import (
	"blog-backend/domain"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	UserUseCase domain.IUserUseCase
}

func NewUserController(uu domain.IUserUseCase) *UserController {
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
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User profile updated successfully."})
}

func (uc *UserController) PromoteUser(c *gin.Context) {
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

func (uc *UserController) DemoteUser(c *gin.Context) {
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
func (uc *UserController) GetUsers(c *gin.Context) {
	p := c.Query("page")
	l := c.Query("limit")
	page, err := strconv.Atoi(p)
	if err != nil || p == "" {
		page = 1
	}
	limit, err := strconv.Atoi(l)
	if err != nil || l == "" {
		limit = 10
	}
	users, total, err := uc.UserUseCase.GetUsers(c, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch Users."})
		return
	}
	totalPages := (total + int64(limit) - 1) / int64(limit)

	c.JSON(http.StatusOK, gin.H{"users": users, "total": totalPages, "page": page, "limit": limit})
}

func (uc *UserController) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	err := uc.UserUseCase.DeleteUser(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete blog."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Blog deleted successfully."})
}
