package infrastructure

import (
	"blog-backend/domain"
	"fmt"
	"net/smtp"
)

type emailServices struct{
	EmailAccount string
	AppPassword  string
}

func NewEmailServices(emailAccount, appPassword string) domain.IEmailServices {
	return &emailServices{
		EmailAccount: emailAccount,
		AppPassword: appPassword,
	}
}

func (es *emailServices) SendActivationEmail(email, activationToken string) error {
    from := es.EmailAccount
    password := es.AppPassword

    smtpHost := "smtp.gmail.com"
    smtpPort := "587"

    activationLink := fmt.Sprintf("http://localhost:3000/api/auth/activate?token=%s", activationToken)

    subject := "Activate Your Account"
    body := fmt.Sprintf(`
        <html>
            <body>
                <h2>Welcome to Blog Backend 🎉</h2>
                <p>Click the button below to activate your account:</p>
                <a href="%s" style="
                    background-color: #4CAF50;
                    color: white;
                    padding: 10px 20px;
                    text-decoration: none;
                    display: inline-block;
                    border-radius: 5px;
                ">Activate Account</a>
                <p>If the button doesn't work, copy and paste this link in your browser:</p>
                <p>%s</p>
            </body>
        </html>
    `, activationLink, activationLink)

    // Combine headers + body
    msg := []byte(fmt.Sprintf("Subject: %s\r\n", subject) +
        "MIME-version: 1.0;\r\n" +
        "Content-Type: text/html; charset=\"UTF-8\";\r\n\r\n" +
        body)

    // Auth and send
    auth := smtp.PlainAuth("", from, password, smtpHost)
    err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{email}, msg)
    return err
}

func (es *emailServices) SendPasswordResetEmail(email, resetToken string) error {
    from := es.EmailAccount
    password := es.AppPassword

    smtpHost := "smtp.gmail.com"
    smtpPort := "587"

    resetLink := fmt.Sprintf("http://localhost:3000/api/auth/reset-password?token=%s", resetToken)

    subject := "Reset Your Password"
    body := fmt.Sprintf(`
        <html>
            <body>
                <h2>Password Reset Request</h2>
                <p>Click the button below to reset your password:</p>
                <a href="%s" style="
                    background-color: #f44336;
                    color: white;
                    padding: 10px 20px;
                    text-decoration: none;
                    display: inline-block;
                    border-radius: 5px;
                ">Reset Password</a>
                <p>If the button doesn't work, copy and paste this link in your browser:</p>
                <p>%s</p>
            </body>
        </html>
    `, resetLink, resetLink)

    // Combine headers + body
    msg := []byte(fmt.Sprintf("Subject: %s\r\n", subject) +
        "MIME-version: 1.0;\r\n" +
        "Content-Type: text/html; charset=\"UTF-8\";\r\n\r\n" +
        body)

    // Auth and send
    auth := smtp.PlainAuth("", from, password, smtpHost)
    err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{email}, msg)
    return err
}