package controller

import "blog-backend/domain"

type AuthController struct {
	AuthUseCase domain.IAuthUseCase
}

func NewAuthController(au domain.IAuthUseCase) *AuthController {
	return &AuthController{
		AuthUseCase: au,
	}
}