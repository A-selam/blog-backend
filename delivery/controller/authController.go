package controller

import (
	"blog-backend/domain"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type AuthController struct {
	AuthUseCase domain.IAuthUseCase
	googleConfig *oauth2.Config
}

func NewAuthController(au domain.IAuthUseCase, googleConfig *oauth2.Config) *AuthController {
	return &AuthController{
		AuthUseCase: au,
		googleConfig: googleConfig,
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
	cook, err := c.Cookie("refresh_token")
	fmt.Println("Refresh Token:", cook, err)
	fmt.Println("Refresh Token2:", tokenPair.RefreshToken)

	c.JSON(http.StatusOK, gin.H{
		"User":      loginResponseFromDomain(user),
		"TokenPair": tokenPair.AccessToken,
	})
}

func (ac *AuthController) RefreshToken(c *gin.Context) {
	// 1. Get the refresh token from the cookie
	refreshToken, err := c.Cookie("refresh_token")
	fmt.Println("Refresh Token:", refreshToken, err)
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
		"Token": tokenPair.AccessToken,
	})
}

func (ac *AuthController) GoogleLogin(c *gin.Context){
	url := ac.googleConfig.AuthCodeURL("state-secret-token", oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (ac *AuthController) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing Code"})
	}

	token, err := ac.googleConfig.Exchange(c, code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error":"Failed to exchange token"})
	}

	client := ac.googleConfig.Client(c, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user info"})
		return
	}
	defer resp.Body.Close()

	var googleUser struct{
		Id      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode user info"})
		return
	}

	user, err := ac.AuthUseCase.FindOrCreateGoogleUser(c, googleUser.Email, googleUser.Name, googleUser.Picture, googleUser.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process Google login"})
		return
	}

	tokenPair, err := ac.AuthUseCase.IssueTokenPair(c, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to issue token pair"})
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
        "User": loginResponseFromDomain(user),
        "Token": tokenPair.AccessToken,
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
	Status   string `json:"status"`
}

type signUpDetailResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Status   string `json:"status"`
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
		Status:   string(u.Status),
	}
}

func signUpDetailResponseFromDomain(u *domain.User) signUpDetailResponse {
	return signUpDetailResponse{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		Role:     string(u.Role),
		Status:   string(u.Status),
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
		GoogleID: "",
		Username: s.Username,
		Email: s.Email,
		PasswordHash: s.Password,
		Role: domain.RegularUser,
		Status: domain.Inactive,
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
		return fmt.Errorf("invalid email format")
	}
	return nil
}

func (ac *AuthController)Logout(c *gin.Context ) {
	RefreshToken, err := c.Cookie("refresh_token")
	
	fmt.Println("Refresh Token:", RefreshToken, err)
	if err != nil || RefreshToken == "" {
		c.JSON(400, gin.H{"error": "Refresh token missing"})
		return
	}
	c.SetCookie("refresh_token", "", -1, "/", "", true, true)

	err = ac.AuthUseCase.Logout(c, RefreshToken)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to logout user."})
		return
	}
	c.JSON(200, gin.H{"message": "User logged out successfully!"})
}

func (ac *AuthController)ForgotPassword(c *gin.Context) {
	var request struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request: email is required"})
		return
	}
	 err := ac.AuthUseCase.ForgotPassword(c, request.Email)
	if err != nil {	
		c.JSON(500, gin.H{"error": "Failed to process forgot password request."})
		return	

	}
	c.JSON(200, gin.H{"message": "Password reset email sent successfully. Please check your inbox."})
}
func (ac *AuthController) ResetPassword(c *gin.Context) {
	token:= c.Query("token")
	if token == "" {	
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reset token is required"})
		return
	}

    var request struct {
        Password string `json:"password" binding:"required,min=8"`
    }
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: token and password (minimum 8 characters) are required"})
        return
    }

    err := ac.AuthUseCase.ResetPassword(c, token, request.Password)
    if err != nil {
        switch err {
        case domain.ErrTokenNotFound:
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or unknown reset token"})
        case domain.ErrTokenUsed:
            c.JSON(http.StatusBadRequest, gin.H{"error": "Reset token has already been used"})
        case domain.ErrTokenExpired:
            c.JSON(http.StatusBadRequest, gin.H{"error": "Reset token has expired"})
        default:
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process reset password"})
        }
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}