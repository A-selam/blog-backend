package domain

import "context"

type GeneratePrompt struct {
	Title string
	MinLength int
	MaxLength int
}

type RefinePrompt struct {
	Content string
}

type IGeminiUseCase interface {
	GenerateContent(ctx context.Context, prompt GeneratePrompt) (string, error)
	RefineContent(ctx context.Context, prompt RefinePrompt) (string, error)
	GenerateSkeleton(ctx context.Context, title string) (string, error)
}

type IGeminiService interface {
	GenerateContent(fullPrompt string) (string, error)
	RefineContent(prompt string) (string, error)
	GenerateSkeleton(prompt string) (string, error)
	GenerateTags(prompt string) ([]string, error)
}