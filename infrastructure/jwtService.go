package infrastructure

import (
	"blog-backend/domain"
	"time"
	"fmt"
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
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		return "", "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(string)
		if !ok {
			return "", "", fmt.Errorf("invalid user id in token")
		}
		role, ok := claims["role"].(string)
		if !ok {
			return "", "", fmt.Errorf("invalid role in token")
		}
		return userID, role, nil
	}

	return "", "", fmt.Errorf("invalid token")
}