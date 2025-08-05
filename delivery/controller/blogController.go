package controller

import (
	"blog-backend/domain"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type BlogController struct {
	BlogUseCase domain.IBlogUseCase
}

func NewBlogController(bu domain.IBlogUseCase) *BlogController {
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

func (bc *BlogController) ListBlogs(c *gin.Context) {
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
	blogs, total, err := bc.BlogUseCase.ListBlogs(c, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch blogs."})
		return
	}
	totalPages := (total + int64(limit) - 1) / int64(limit)

	c.JSON(http.StatusOK, gin.H{"blogs": blogs, "total": totalPages, "page": page, "limit": limit})
}

func (bc *BlogController) GetBlog(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(400, gin.H{"error": "Blog ID is required."})
	}
	blog, err := bc.BlogUseCase.GetBlog(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch blog."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"blog": blog})
}

func (bc *BlogController) UpdateBlog(c *gin.Context) {
	id := c.Param("id")
	var updates BlogUpdateDTO
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(400, gin.H{"error": "Invalid entry"})
	}
	if id == "" {
		c.JSON(400, gin.H{"error": "Blog ID is required."})
	}
	updatesMap := map[string]interface{}{}
	if updates.Title != nil {
		updatesMap["title"] = *updates.Title
	}
	if updates.Content != nil {
		updatesMap["content"] = *updates.Content
	}
	if updates.Tags != nil {
		updatesMap["tags"] = *updates.Tags
	}
	userID, exists := c.Get("x-user-id")
	if !exists {
		c.JSON(400, gin.H{"error": "No user ID found"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(400, gin.H{"error": "Invalid user ID type"})
		return
	}
	if len(updatesMap) == 0 {
		c.JSON(400, gin.H{"error": "No fields to update."})
		return
	}

	err := bc.BlogUseCase.UpdateBlog(c, id, userIDStr, updatesMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update blog."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Blog updated successfully."})

}
func (bc *BlogController) RemoveReaction(c *gin.Context) {
	blogID := c.Param("id")
	userID, exists := c.Get("x-user-id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	err := bc.BlogUseCase.RemoveReaction(c, blogID, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove reaction", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reaction removed successfully"})
}

func (bc * BlogController) CreateComment(c *gin.Context){
	blogID := c.Param("id")
	userID, exists := c.Get("x-user-id")
	if !exists{
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
	}
	if blogID == ""{
		c.JSON(400, gin.H{"error": "Invalid Request. Blog ID is required"})
		return

	}
	type CommentDTO struct {
    Comment string `json:"comment" binding:"required"`
}
	var commentDTO CommentDTO
	if err := c.ShouldBindJSON(&commentDTO); err != nil{
		c.JSON(400, gin.H{"error": "Invalid Request. Comment is required"})
		return
	}
	_, err := bc.BlogUseCase.AddComment(c, blogID, userID.(string), commentDTO.Comment)
	if err !=nil{
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to add you comment"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "comment added successfully!"})

}

func (bc *BlogController) ListAllComments(c *gin.Context){
	blogID := c.Param("id")
	if blogID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Blog ID is required"})
		return
	}

	comments, err := bc.BlogUseCase.GetComments(c, blogID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch comments", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"comments": comments})

}

func (bc * BlogController) CreateComment(c *gin.Context){
	blogID := c.Param("id")
	userID, exists := c.Get("x-user-id")
	if !exists{
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
	}
	if blogID == ""{
		c.JSON(400, gin.H{"error": "Invalid Request. Blog ID is required"})
		return

	}
	type CommentDTO struct {
    Comment string `json:"comment" binding:"required"`
}
	var commentDTO CommentDTO
	if err := c.ShouldBindJSON(&commentDTO); err != nil{
		c.JSON(400, gin.H{"error": "Invalid Request. Comment is required"})
		return
	}
	_, err := bc.BlogUseCase.AddComment(c, blogID, userID.(string), commentDTO.Comment)
	if err !=nil{
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to add you comment"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "comment added successfully!"})

}

func (bc *BlogController) ListAllComments(c *gin.Context){
	blogID := c.Param("id")
	if blogID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Blog ID is required"})
		return
	}

	comments, err := bc.BlogUseCase.GetComments(c, blogID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch comments", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"comments": comments})

}



func (bc *BlogController) LikeBlog(c *gin.Context) {
	blogID := c.Param("id")
	if blogID == "" {
		c.JSON(400, gin.H{"error": "Blog ID is required."})
		return
	}

	userID, exists := c.Get("x-user-id")
	if !exists {
		c.JSON(400, gin.H{"error": "User ID not found in context"})
		return
	}

	err := bc.BlogUseCase.AddReaction(c, blogID, userID.(string), string(domain.Like))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to add reaction."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reaction added successfully."})
}

func (bc *BlogController) DislikeBlog(c *gin.Context) {
	blogID := c.Param("id")
	if blogID == "" {
		c.JSON(400, gin.H{"error": "Blog ID is required."})
		return
	}

	userID, exists := c.Get("x-user-id")
	if !exists {
		c.JSON(400, gin.H{"error": "User ID not found in context"})
		return
	}

	err := bc.BlogUseCase.AddReaction(c, blogID, userID.(string), string(domain.Dislike))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add reaction."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reaction added successfully."})
}

type BlogDTO struct {
	Title   string   `json:"title" binding:"required"`
	Content string   `json:"content" binding:"required"`
	Tags    []string `json:"tags" binding:"required"`
}

func DtoToDomain(blogDTO *BlogDTO, authorID string) *domain.Blog {
	return &domain.Blog{
		Title:    blogDTO.Title,
		Content:  blogDTO.Content,
		AuthorID: authorID,
		Tags:     blogDTO.Tags,	
		ViewCount:    0,
		LikeCount:    0,
		DislikeCount: 0,
		CommentCount: 0,
	}
}
type BlogUpdateDTO struct {
	Title   *string   `json:"title,omitempty"`
	Content *string   `json:"content,omitempty"`
	Tags    *[]string `json:"tags,omitempty"`
}