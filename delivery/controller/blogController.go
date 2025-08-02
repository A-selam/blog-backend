package controller

import (
	"blog-backend/domain"
	"net/http"

	"github.com/gin-gonic/gin"
)

type BlogController struct {
	BlogUseCase domain.IBlogUseCase
}

func NewBlogController(bu domain.IBlogUseCase) *BlogController{
	return &BlogController{
		BlogUseCase: bu,
	}
}

func (bc *BlogController) CreateBlog(c *gin.Context) {
	userID, exists := c.Get("x-user-id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	var blogDTO BlogDTO
	if err := c.ShouldBindJSON(&blogDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	blog := DtoToDomain(&blogDTO, userID.(string))
	createdBlog, err := bc.BlogUseCase.CreateBlog(c, blog)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create blog", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"blog": createdBlog})
}

// DTOs

type BlogDTO struct {
	Title  string   `json:"title" binding:"required"`
	Content string   `json:"content" binding:"required"`
	Tags   []string `json:"tags" binding:"required"`
}

func DtoToDomain(blogDTO *BlogDTO, authorID string) *domain.Blog {
	return &domain.Blog{
		Title:    blogDTO.Title,
		Content:  blogDTO.Content,
		AuthorID: authorID,
		Tags:     blogDTO.Tags,
	}
}