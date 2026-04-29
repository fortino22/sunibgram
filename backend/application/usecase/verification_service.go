package usecase

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/infrastructure/ratelimit"
	"golang.org/x/crypto/bcrypt"
)

type VerificationService struct {
	verificationRepo repository.IVerificationRepository
	emailService     service.IEmailService
	sendLimiter      *ratelimit.RateLimiter
	verifyLimiter    *ratelimit.RateLimiter
}

func NewVerificationService(
	verificationRepo repository.IVerificationRepository,
	emailService service.IEmailService,
) service.IVerificationService {
	return &VerificationService{
		verificationRepo: verificationRepo,
		emailService:     emailService,
		sendLimiter:      ratelimit.New(3, 10*time.Minute),
		verifyLimiter:    ratelimit.New(3, 10*time.Minute),
	}
}

func (s *VerificationService) SendCode(email string) error {
	if err := model.ValidateVerificationEmail(email); err != nil {
		return err
	}

	if err := s.sendLimiter.Allow(email); err != nil {
		return err
	}

	_ = s.verificationRepo.DeleteByEmail(email)

	plain := fmt.Sprintf("%06d", rand.Intn(1000000))

	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to process code")
	}

	record := &model.VerificationCode{
		Email:     email,
		Code:      string(hashed),
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	if err := s.verificationRepo.Save(record); err != nil {
		return err
	}

	subject := "Your Verification Code"
	body := fmt.Sprintf(`
		<div style="font-family:sans-serif;max-width:480px;margin:auto">
			<h2>Verification Code</h2>
			<p>Use the code below to verify your email. It expires in <strong>10 minutes</strong>.</p>
			<div style="font-size:32px;font-weight:bold;letter-spacing:8px;padding:16px;background:#f4f4f4;text-align:center;border-radius:8px">%s</div>
			<p style="color:#888;font-size:12px;margin-top:16px">If you didn't request this, you can safely ignore this email.</p>
		</div>`, plain)

	return s.emailService.SendMail(email, subject, body)
}

func (s *VerificationService) VerifyCode(email, code string) error {
	if err := s.verifyLimiter.Allow(email); err != nil {
		return err
	}

	record, err := s.verificationRepo.FindLatestByEmail(email)
	if err != nil {
		return errors.New("invalid or expired code")
	}

	if time.Now().After(record.ExpiresAt) {
		return errors.New("code has expired")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(record.Code), []byte(code)); err != nil {
		return errors.New("invalid code")
	}

	return s.verificationRepo.MarkUsed(record.ID)
}
