package usecase

import (
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
