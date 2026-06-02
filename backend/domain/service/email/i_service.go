package emailsvciface

import "context"

type IEmailService interface {
	SendMail(to, subject, body string) error
	QueueMail(ctx context.Context, to, subject, body string) error
	StartWorker(ctx context.Context)
	SendWelcomeOTP(to, code string) error
	SendPasswordResetOTP(to, code string) error
	SendLoginAlert(to, yesURL, noURL, device, ip string) error
	SendChangePasswordOTP(to, code string) error
}
