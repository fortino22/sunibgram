package service

type IEmailService interface {
	SendMail(to, subject, body string) error
}
