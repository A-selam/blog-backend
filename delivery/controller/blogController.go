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

func (bc *BlogController) DeleteBlog(c *gin.Context) {
	blogID := c.Param("id")
	userID, _ := c.Get("x-user-id")
	userRole, _ := c.Get("x-user-role")

	isAuthor, err := bc.BlogUseCase.IsBlogAuthor(c, blogID, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify author."})
		return
	}

	if !isAuthor && userRole.(string) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to delete this blog."})
		return
	}

	err = bc.BlogUseCase.DeleteBlog(c, blogID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete blog."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Blog deleted successfully."})
}

func (bc *BlogController) SearchBlogs(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required."})
		return
	}

	blogs, err := bc.BlogUseCase.SearchBlogs(c, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search blogs."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"blogs": blogs})
}

func (bc *BlogController) GetBlogsByUserID(c *gin.Context) {
	userID := c.Param("id")

	blogs, err := bc.BlogUseCase.GetBlogsByUserID(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user's blogs."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"blogs": blogs})
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