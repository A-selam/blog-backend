package controller

import "blog-backend/domain"

type BlogController struct {
	BlogUseCase domain.IBlogUseCase
}

func NewBlogController(bu domain.IBlogUseCase) *BlogController{
	return &BlogController{
		BlogUseCase: bu,
	}
}