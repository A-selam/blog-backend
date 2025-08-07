package domain

type IEmailServices interface {
	SendActivationEmail(email, activationToken string) error
	SendPasswordResetEmail(email, resetToken string) error
}