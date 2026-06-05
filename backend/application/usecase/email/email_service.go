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

// loginLogoSVG is the Sunibgram bolt mark (from frontend/components/ui/Logo.tsx)
// flattened to a plain inline SVG. Rendered in supporting mail clients
// (Apple Mail etc.); Gmail strips inline SVG, leaving the wordmark as fallback.
const loginLogoSVG = `<svg width="24" height="24" viewBox="0 0 384 512" fill="#0f1419" style="display:block"><path d="M236 0 L40 300 L190 300 L150 512 L344 224 L190 224 Z"/></svg>`

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
	return s.QueueMail(context.Background(), to, subject, body)
}

// SendLoginAlert notifies the account owner that a new sign-in just happened.
// Notify-only: it carries no action links. Destructive links here get prefetched
// by webmail/antivirus scanners, which would silently revoke the just-created
// session. If the sign-in is unrecognized, the owner resets their password (which
// revokes all sessions) via the app.
func (s *EmailService) SendLoginAlert(to, device, ip string) error {
	subject := "Sunibgram — New Sign-In to Your Account"
	const font = "-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif"
	body := fmt.Sprintf(`
		<table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="background:#ffffff;margin:0;padding:24px 0">
		<tr><td align="center">
		<table role="presentation" width="480" cellpadding="0" cellspacing="0" style="max-width:480px;width:100%%;background:#ffffff;border:1px solid #eff3f4;border-radius:16px">
		<tr><td style="padding:32px 32px 28px 32px;font-family:%[1]s">
			<table role="presentation" cellpadding="0" cellspacing="0" style="margin-bottom:24px"><tr>
				<td style="vertical-align:middle;padding-right:8px;line-height:0;color:#0f1419">%[4]s</td>
				<td style="vertical-align:middle;font-size:20px;font-weight:800;color:#0f1419;letter-spacing:-0.02em">Sunibgram</td>
			</tr></table>
			<div style="font-size:23px;font-weight:800;color:#0f1419;letter-spacing:-0.02em;line-height:1.3;margin-bottom:8px">New sign-in detected</div>
			<p style="font-size:15px;color:#536471;line-height:1.5;margin:0 0 20px 0">Someone just signed in to your Sunibgram account. If this was you, no action is needed.</p>
			<table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="background:#f7f9f9;border-radius:12px;margin:0 0 24px 0">
			<tr><td style="padding:14px 16px;font-family:%[1]s;font-size:14px;color:#0f1419;line-height:1.6">
				<span style="color:#536471">Device</span><br>%[2]s<br>
				<span style="color:#536471">IP address</span><br>%[3]s
			</td></tr></table>
			<p style="font-size:13px;color:#536471;line-height:1.5;margin:0">If you don't recognize this sign-in, <strong style="color:#0f1419">reset your password</strong> from the Sunibgram app to secure your account and sign out every device.</p>
		</td></tr></table>
		</td></tr></table>`, font, device, ip, loginLogoSVG)
	return s.QueueMail(context.Background(), to, subject, body)
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
	return s.QueueMail(context.Background(), to, subject, body)
}

func (s *EmailService) SendChangePasswordOTP(to, code string) error {
	subject := "Sunibgram — Password Change Code"
	body := fmt.Sprintf(`
		<div style="font-family:sans-serif;max-width:480px;margin:auto">
			<h2>Confirm Password Change</h2>
			<p>Use the code below to change your password. It expires in <strong>10 minutes</strong>.</p>
			<div style="font-size:32px;font-weight:bold;letter-spacing:8px;padding:16px;background:#f4f4f4;text-align:center;border-radius:8px">%s</div>
			<p style="color:#888;font-size:12px;margin-top:16px">If you didn't request a password change, please secure your account immediately.</p>
		</div>`, code)
	return s.QueueMail(context.Background(), to, subject, body)
}
