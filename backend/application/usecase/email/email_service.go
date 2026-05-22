package emailuc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	emailsvciface "github.com/sunibgram/domain/service/email"
	"github.com/sunibgram/infrastructure/cache"
	"github.com/sunibgram/infrastructure/email"
)

const emailQueueKey = "email:queue"

type emailJob struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type EmailService struct {
	client *email.SMTPClient
	cache  *cache.Cache
}

func NewEmailService(client *email.SMTPClient, redisCache *cache.Cache) emailsvciface.IEmailService {
	return &EmailService{client: client, cache: redisCache}
}

func (s *EmailService) SendMail(to, subject, body string) error {
	return s.client.SendMail(to, subject, body)
}

func (s *EmailService) QueueMail(ctx context.Context, to, subject, body string) error {
	return s.cache.RPush(ctx, emailQueueKey, emailJob{To: to, Subject: subject, Body: body})
}

func (s *EmailService) StartWorker(ctx context.Context) {
	for {
		payload, err := s.cache.BLPop(ctx, emailQueueKey, 0)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("email worker dequeue error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		var job emailJob
		if err := json.Unmarshal([]byte(payload), &job); err != nil {
			log.Printf("email worker decode error: %v", err)
			continue
		}
		if err := s.SendMail(job.To, job.Subject, job.Body); err != nil {
			log.Printf("email worker send error to=%s: %v", job.To, err)
		}
	}
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

func (s *EmailService) SendLoginOTP(to, code string) error {
	subject := "Sunibgram — Login Verification Code"
	body := fmt.Sprintf(`
		<div style="font-family:sans-serif;max-width:480px;margin:auto">
			<h2>Login Verification</h2>
			<p>Use the code below to complete your login. It expires in <strong>10 minutes</strong>.</p>
			<div style="font-size:32px;font-weight:bold;letter-spacing:8px;padding:16px;background:#f4f4f4;text-align:center;border-radius:8px">%s</div>
			<p style="color:#888;font-size:12px;margin-top:16px">If you didn't attempt to log in, please change your password immediately.</p>
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
