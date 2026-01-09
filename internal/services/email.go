package services

import (
	"fmt"

	"github.com/tresor/password-manager/internal/config"
	"gopkg.in/gomail.v2"
)

type EmailService struct {
	config *config.EmailConfig
}

func NewEmailService(cfg *config.EmailConfig) *EmailService {
	return &EmailService{config: cfg}
}

func (s *EmailService) SendShareNotification(recipientEmail, title, shareURL string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.config.From)
	m.SetHeader("To", recipientEmail)
	m.SetHeader("Subject", "Password Shared With You")

	body := fmt.Sprintf(`
        <html>
        <body>
            <h2>A password has been shared with you</h2>
            <p>Someone has shared <strong>%s</strong> with you.</p>
            <p>Click the link below to access it:</p>
            <p><a href="%s">Access Shared Password</a></p>
            <p>This link may expire or have limited views. Make sure to save the information if needed.</p>
            <br>
            <p><em>SecureVault - Your Password Manager</em></p>
        </body>
        </html>
    `, title, shareURL)

	m.SetBody("text/html", body)

	d := gomail.NewDialer(s.config.Host, s.config.Port, s.config.Username, s.config.Password)

	return d.DialAndSend(m)
}

func (s *EmailService) SendWelcomeEmail(email string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.config.From)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Welcome to SecureVault")

	body := `
        <html>
        <body>
            <h2>Welcome to SecureVault!</h2>
            <p>Thank you for signing up. Your account has been created successfully.</p>
            <p>Start securing your passwords today!</p>
            <br>
            <p><em>SecureVault Team</em></p>
        </body>
        </html>
    `

	m.SetBody("text/html", body)

	d := gomail.NewDialer(s.config.Host, s.config.Port, s.config.Username, s.config.Password)

	return d.DialAndSend(m)
}

func (s *EmailService) Send2FACode(email, code string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.config.From)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Your 2FA Code")

	body := fmt.Sprintf(`
        <html>
        <body>
            <h2>Two-Factor Authentication Code</h2>
            <p>Your verification code is: <strong style="font-size: 24px;">%s</strong></p>
            <p>This code will expire in 10 minutes.</p>
            <br>
            <p><em>SecureVault Security</em></p>
        </body>
        </html>
    `, code)

	m.SetBody("text/html", body)

	d := gomail.NewDialer(s.config.Host, s.config.Port, s.config.Username, s.config.Password)

	return d.DialAndSend(m)
}
