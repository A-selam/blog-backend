package controller

import "blog-backend/domain"

type UserController struct {
	UserUseCase domain.IUserUseCase
}

func NewUserController(uu domain.IUserUseCase) *UserController{
	return &UserController{
		UserUseCase: uu,
	}
}
