package controller

import (
	"blog-backend/domain"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GeminiController struct {
	geminiUseCase domain.IGeminiUseCase
}

func NewGeminiController(geminiUseCase domain.IGeminiUseCase) *GeminiController {
	return &GeminiController{
		geminiUseCase: geminiUseCase,
	}
}

func (gc *GeminiController) GenerateContent(c *gin.Context){
	var generatePrompt GeneratePromptDTO
	err := c.ShouldBindJSON(&generatePrompt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid request body"})
		return 
	}
	
	if generatePrompt.Title == ""{
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title cannot be empty."})
		return
	}

	if len(generatePrompt.Title) < 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title must be at least 10 characters long."})
		return
	}

	prompt := GeneratePromptDTOToDomain(&generatePrompt)
	response, err := gc.geminiUseCase.GenerateContent(c, *prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Failed to generate response."})
	}

	c.JSON(http.StatusOK, gin.H{"Response": response})
}

func (gc *GeminiController) RefineContent(c *gin.Context) {
	var refinePrompt PromptDTO
	err := c.ShouldBindJSON(&refinePrompt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid request body"})
		return 
	}
	
	if refinePrompt.Content == ""{
		c.JSON(http.StatusBadRequest, gin.H{"error": "Content cannot be empty."})
		return
	}

	prompt := RefineContentDTOToDomain(&refinePrompt)
	response, err := gc.geminiUseCase.RefineContent(c, *prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Failed to generate response."})
	}

	c.JSON(http.StatusOK, gin.H{"Response": response})
}

func (gc *GeminiController) GenerateSkeleton(c *gin.Context) {
	var title PromptDTO 
	err := c.ShouldBindJSON(&title)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid request body"})
		return 
	}
	
	if title.Content == ""{
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title cannot be empty."})
		return
	}

	if len(title.Content) < 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title must be at least 10 characters long."})
		return
	}

	prompt := title.Content
	response, err := gc.geminiUseCase.GenerateSkeleton(c, prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Failed to generate response."})
	}

	c.JSON(http.StatusOK, gin.H{"Response": response})
}

type PromptDTO struct {
	Content string `json:"content"`
}

type GeneratePromptDTO struct {
	Title string `json:"title"`
	MinLength int `json:"min-length"`
	MaxLength int `json:"max-length"`
}

func GeneratePromptDTOToDomain(d *GeneratePromptDTO) *domain.GeneratePrompt {
	return &domain.GeneratePrompt{
		Title: d.Title,
		MinLength: d.MinLength,
		MaxLength: d.MaxLength,
	}
}

func RefineContentDTOToDomain(d *PromptDTO) *domain.RefinePrompt {
	return &domain.RefinePrompt{
		Content: d.Content,
	}
}