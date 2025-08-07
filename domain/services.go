package domain

type IEmailServices interface {
	SendActivationEmail(email, activationToken string) error
}