package usecase

import (
	"blog-backend/domain"
	"context"
	"fmt"
)

type geminiUseCase struct {
	geminiServices domain.IGeminiService
}

func NewGeminiUsecase(geminiServices domain.IGeminiService) domain.IGeminiUseCase {
	return &geminiUseCase{
		geminiServices: geminiServices,
	}
}

func (gu *geminiUseCase) GenerateContent(ctx context.Context, prompt domain.GeneratePrompt) (string, error) {
	if prompt.MinLength > prompt.MaxLength {
		temp := prompt.MaxLength
		prompt.MaxLength = prompt.MinLength
		prompt.MinLength = temp
	}

	if prompt.MinLength < 300 {
		prompt.MinLength = 300
	}
	if prompt.MaxLength < 600 {
		prompt.MaxLength = 600
	}
	
	if prompt.MinLength > 1000 {
		prompt.MinLength = 1000
	}
	if prompt.MaxLength > 1500 {
		prompt.MaxLength = 1500
	}

	fullPrompt := fmt.Sprintf(`Generate a complete blog post about "%s". The post should be " %d "-"%d" words, include a title, introduction, 3-5 main sections with detailed content, and a conclusion. Write in a clear, engaging, and informative tone suitable for a general audience. Return the content in markdown format with proper line breaks (use actual line breaks, not escaped "\n\n" or other escape sequences). Ensure each section is separated by a blank line for correct markdown rendering.`, prompt.Title, prompt.MinLength, prompt.MaxLength)

	response, err := gu.geminiServices.GenerateContent(fullPrompt)
	if err != nil {
		return "", err
	}
	
	return response, nil
}

func (gu *geminiUseCase) RefineContent(ctx context.Context, prompt domain.RefinePrompt) (string, error) {
	fullPrompt := fmt.Sprintf( `Refine the following blog post content to improve clarity, grammar, and style while maintaining the original meaning. Return the refined content in plain text: "%s"`, prompt.Content)

	response, err := gu.geminiServices.RefineContent(fullPrompt)
	if err != nil {
		return "", err
	}

	return response, nil
}

func (gu *geminiUseCase) GenerateSkeleton(ctx context.Context, title string) (string, error) {
	fullPrompt := fmt.Sprintf(`Generate a blog post skeleton for the topic "%s". The skeleton should include a title, introduction, 3-5 main section headings with 1-2 sentence descriptions, and a conclusion. Do not write a full blog post; provide an outline that the user can fill in with their ideas. Return the skeleton in markdown format with proper line breaks (use actual line breaks, not escaped "\n\n" or other escape sequences). Ensure each section is separated by a blank line for correct markdown rendering.`, title)

	response, err := gu.geminiServices.GenerateSkeleton(fullPrompt)
	if err != nil {
		return "", err
	}

	return response, nil
}