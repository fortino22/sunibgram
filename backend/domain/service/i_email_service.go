package service

type IEmailService interface {
	SendMail(to, subject, body string) error
	SendWelcomeOTP(to, code string) error
	SendPasswordResetOTP(to, code string) error
	SendLoginOTP(to, code string) error
}
