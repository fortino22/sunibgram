package usecase

import (
	"fmt"

	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/infrastructure/email"
)

type EmailService struct {
	client *email.SMTPClient
}

func NewEmailService(client *email.SMTPClient) service.IEmailService {
	return &EmailService{client: client}
}

func (s *EmailService) SendMail(to, subject, body string) error {
	return s.client.SendMail(to, subject, body)
}

func (s *EmailService) SendWelcomeOTP(to, code string) error {
	subject := "Welcome to Sunibgram — Verify Your Email"
	body := fmt.Sprintf(`
		<div style="font-family:sans-serif;max-width:480px;margin:auto">
			<h2>Welcome to Sunibgram!</h2>
			<p>Use the code below to verify your email and complete registration. It expires in <strong>10 minutes</strong>.</p>
			<div style="font-size:32px;font-weight:bold;letter-spacing:8px;padding:16px;background:#f4f4f4;text-align:center;border-radius:8px">%s</div>
			<p style="color:#888;font-size:12px;margin-top:16px">If you didn't create an account, you can safely ignore this email.</p>
		</div>`, code)
	return s.client.SendMail(to, subject, body)
}

func (s *EmailService) SendPasswordResetOTP(to, code string) error {
	subject := "Sunibgram — Password Reset Code"
	body := fmt.Sprintf(`
		<div style="font-family:sans-serif;max-width:480px;margin:auto">
			<h2>Password Reset</h2>
			<p>Use the code below to reset your password. It expires in <strong>10 minutes</strong>.</p>
			<div style="font-size:32px;font-weight:bold;letter-spacing:8px;padding:16px;background:#f4f4f4;text-align:center;border-radius:8px">%s</div>
			<p style="color:#888;font-size:12px;margin-top:16px">If you didn't request a password reset, you can safely ignore this email.</p>
		</div>`, code)
	return s.client.SendMail(to, subject, body)
}
