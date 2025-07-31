package infrastructure

import "blog-backend/domain"

type BcryptPasswordService struct{}

func NewPasswordService() domain.IPasswordService {
	return &BcryptPasswordService{}
}

func (s *BcryptPasswordService) HashPassword(password string) (string, error) {
	// TODO: Implement the password hashing functionality
	return "hashedPassword", nil
}

func (s *BcryptPasswordService) ComparePassword(hashedPassword, plainPassword string) error {
	// TODO: Implement the password comparison function
	return nil
}
