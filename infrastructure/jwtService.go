package infrastructure

import (
	"blog-backend/domain"
	"time"

	"github.com/golang-jwt/jwt"
)

type jWTService struct {
	secret []byte
}

func NewJWTService(secret string) domain.IJWTService {
	return &jWTService{secret: []byte(secret)}
}

func (j *jWTService) GenerateToken(userID,username, email, role string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"username": username,
		"email":    email,
		"role":     role,
		"exp":      time.Now().Add(30 * time.Minute).Unix(),
	})

	return token.SignedString([]byte(j.secret))
}


func (s *jWTService) ParseToken(tokenString string) (string, string, error) {
	// TODO: Validate and parse token
	return "", "", nil
}