package infrastructure

import (
	"blog-backend/domain"
	"context"
	"log"
	"strings"

	"google.golang.org/genai"
)

type geminiServices struct {
	client *genai.Client
}

func NewGeminiService(apiKey string) domain.IGeminiService {
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatal("Failed to create gemini client.")
		return nil
	}
	return &geminiServices{
		client: client,
	}
}

func (gs *geminiServices) GenerateContent(fullPrompt string) (string, error) {
	resp, err := gs.client.Models.GenerateContent(context.Background(), "gemini-2.5-flash", genai.Text(fullPrompt), nil)

	if err != nil {
		return "", err
	}

	return resp.Text(), nil
}

func (gs *geminiServices) RefineContent(prompt string) (string, error) {
	resp, err := gs.client.Models.GenerateContent(context.Background(), "gemini-2.5-flash", genai.Text(prompt), nil)

	if err != nil {
		return "", err
	}

	return resp.Text(), nil
}

func (gs *geminiServices) GenerateSkeleton(prompt string) (string, error) {
	resp, err := gs.client.Models.GenerateContent(context.Background(), "gemini-2.5-flash", genai.Text(prompt), nil)

	if err != nil {
		return "", err
	}

	return resp.Text(), nil
}

func (gs *geminiServices) GenerateTags(prompt string) ([]string, error) {
	resp, err := gs.client.Models.GenerateContent(context.Background(), "gemini-2.5-flash", genai.Text(prompt), nil)

	if err != nil {
		return []string{}, err
	}

	tags := strings.Split(resp.Text(), ",")
	for i := range(tags){
		tags[i] = strings.TrimSpace(tags[i])
	}

	return tags, nil
}