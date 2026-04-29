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
	userRepo         repository.IUserRepository
	emailService     service.IEmailService
	sendLimiter      *ratelimit.RateLimiter
	verifyLimiter    *ratelimit.RateLimiter
}

func NewVerificationService(
	verificationRepo repository.IVerificationRepository,
	userRepo repository.IUserRepository,
	emailService service.IEmailService,
) service.IVerificationService {
	return &VerificationService{
		verificationRepo: verificationRepo,
		userRepo:         userRepo,
		emailService:     emailService,
		sendLimiter:      ratelimit.New(3, 10*time.Minute),
		verifyLimiter:    ratelimit.New(3, 10*time.Minute),
	}
}

func (s *VerificationService) SendCode(email string, t model.VerificationType) error {
	if err := model.ValidateVerificationEmail(email); err != nil {
		return err
	}
	if err := model.ValidateVerificationType(t); err != nil {
		return err
	}

	exists, err := s.userRepo.ExistsByEmail(email)
	if err != nil {
		return err
	}
	switch t {
	case model.VerificationTypeRegister:
		if exists {
			return errors.New("email already registered")
		}
	case model.VerificationTypeForgotPassword:
		if !exists {
			return errors.New("email not found")
		}
	}

	rateLimitKey := string(t) + ":" + email
	if err := s.sendLimiter.Allow(rateLimitKey); err != nil {
		return err
	}

	_ = s.verificationRepo.DeleteByEmailAndType(email, t)

	plain := fmt.Sprintf("%06d", rand.Intn(1000000))
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to process code")
	}

	record := &model.VerificationCode{
		Email:     email,
		Code:      string(hashed),
		Type:      t,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	if err := s.verificationRepo.Save(record); err != nil {
		return err
	}

	switch t {
	case model.VerificationTypeRegister:
		return s.emailService.SendWelcomeOTP(email, plain)
	case model.VerificationTypeForgotPassword:
		return s.emailService.SendPasswordResetOTP(email, plain)
	default:
		return errors.New("unsupported verification type")
	}
}

func (s *VerificationService) VerifyCode(email, code string, t model.VerificationType) error {
	if err := model.ValidateVerificationType(t); err != nil {
		return err
	}

	rateLimitKey := string(t) + ":" + email
	if err := s.verifyLimiter.Allow(rateLimitKey); err != nil {
		return err
	}

	record, err := s.verificationRepo.FindLatestByEmailAndType(email, t)
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
