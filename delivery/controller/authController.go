package controller

import (
	"blog-backend/domain"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	AuthUseCase domain.IAuthUseCase
}

func NewAuthController(au domain.IAuthUseCase) *AuthController {
	return &AuthController{
		AuthUseCase: au,
	}
}

func (ac *AuthController) Register(c *gin.Context) {
	var signUpDetail signUpDTO
	if err := c.ShouldBindJSON(&signUpDetail); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if signUpDetail.Username == "" || signUpDetail.Password == "" || signUpDetail.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Empty required field."})
		return
	}

	if len(signUpDetail.Username) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username must be at least 3 characters long."})
		return
	}

	if len(signUpDetail.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 8 characters long."})
		return
	}

	if err := emailValidator(signUpDetail.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Simple email regex
	emailRegex := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	if matched := regexp.MustCompile(emailRegex).MatchString(signUpDetail.Email); !matched {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format."})
		return
	}

	user := signUpToDomain(signUpDetail)

	registerdUser, err := ac.AuthUseCase.Register(c, user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	registeredUserResponse := signUpDetailResponseFromDomain(registerdUser)

	c.JSON(http.StatusCreated, gin.H{"User": registeredUserResponse})
}

func (ac *AuthController) Login(c *gin.Context) {
	var loginDetail loginDTO
	if err := c.ShouldBindJSON(&loginDetail); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if loginDetail.Email == "" || loginDetail.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email and password are required."})
		return
	}

	user, tokenPair, err := ac.AuthUseCase.Login(c, loginDetail.Email, loginDetail.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie(
		"refresh_token",
		tokenPair.RefreshToken,
		int(time.Until(tokenPair.ExpiresIn).Seconds()),
		"/",
		"",
		true,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"User":      loginResponseFromDomain(user),
		"TokenPair": tokenPair.AccessToken,
	})
}

func (ac *AuthController) RefreshToken(c *gin.Context) {
	// 1. Get the refresh token from the cookie
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token missing"})
		return
	}

	// 2. Validate and process the refresh token using your usecase/service
	user, tokenPair, err := ac.AuthUseCase.RefreshToken(c, refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	// 3. Set a new refresh token cookie if you rotate tokens (optional, but recommended)
	c.SetCookie(
		"refresh_token",
		tokenPair.RefreshToken,
		int(time.Until(tokenPair.ExpiresIn).Seconds()),
		"/",
		"",
		true,
		true,
	)

	// 4. Return the new access token (and user info if needed)
	c.JSON(http.StatusOK, gin.H{
		"User": loginResponseFromDomain(user),
		"TokenPair": tokenPair.AccessToken,
	})
}

// DTO
type signUpDTO struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type loginDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type signUpDetailResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Bio            string `json:"bio"`
	ProfilePicture string `json:"profile_picture"`
	ContactInfo    string `json:"contact_info"`
}

func loginResponseFromDomain(u *domain.User) loginResponse {
	return loginResponse{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		Role:     string(u.Role),
	}
}

func signUpDetailResponseFromDomain(u *domain.User) signUpDetailResponse {
	return signUpDetailResponse{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		Role:     string(u.Role),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,

		Bio:            u.Bio,
		ProfilePicture: u.ProfilePicture,
		ContactInfo:    u.ContactInfo,
	}
}

func signUpToDomain(s signUpDTO) *domain.User {
	now := time.Now()
	return &domain.User{
		ID: "",
		Username: s.Username,
		Email: s.Email,
		PasswordHash: s.Password,
		Role: domain.RegularUser,
		CreatedAt: now,
		UpdatedAt: now,

		Bio: "",
		ProfilePicture: "",
		ContactInfo: "",
	}
}

func emailValidator(email string) error {
	// Simple email regex
	emailRegex := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	if matched := regexp.MustCompile(emailRegex).MatchString(email); !matched {
		return fmt.Errorf("Invalid email format.")
	}
	return nil
}