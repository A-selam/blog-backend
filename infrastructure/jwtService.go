package infrastructure

import (
	"blog-backend/domain"
)

type jWTService struct {
	secret []byte
}

func NewJWTService(secret string) domain.IJWTService {
	return &jWTService{secret: []byte(secret)}
}

func (j *jWTService) GenerateToken(username, role string) (string, error) {
	// TODO: Generate access token
	return "", nil
}


func (s *jWTService) ParseToken(tokenString string) (string, string, error) {
	// TODO: Validate and parse token
	return "", "", nil
}