package usecase

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	verificationcodemodel "github.com/sunibgram/domain/model/verification_code"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/infrastructure/ratelimit"
)

const otpHashV2Prefix = "v2:"

// otpHMACSecret resolves the HMAC key for OTP hashing. Prefers a dedicated
// OTP_HMAC_SECRET; falls back to JWT_SECRET to avoid forcing a new env var on
// deployers. Last-resort literal matches the existing JWT fallback pattern so
// no environment becomes unbootable; rotate by setting OTP_HMAC_SECRET.
func otpHMACSecret() []byte {
	if s := strings.TrimSpace(os.Getenv("OTP_HMAC_SECRET")); s != "" {
		return []byte(s)
	}
	if s := strings.TrimSpace(os.Getenv("JWT_SECRET")); s != "" {
		return []byte(s)
	}
	return []byte("sunibgram-otp")
}

// hashOTP returns a versioned HMAC-SHA256 hash of the plaintext code. The
// "v2:" prefix discriminates from legacy bcrypt hashes ("$2a$...") still
// present in the verification_codes table from before this migration. Bcrypt
// rows drain naturally via the 10-minute OTP TTL.
func hashOTP(plain string) string {
	h := hmac.New(sha256.New, otpHMACSecret())
	h.Write([]byte(plain))
	return otpHashV2Prefix + hex.EncodeToString(h.Sum(nil))
}

// verifyOTPHash dispatches by hash prefix: v2 → HMAC compare in constant
// time; anything else → bcrypt compare (legacy in-flight codes).
func verifyOTPHash(stored, plain string) bool {
	if strings.HasPrefix(stored, otpHashV2Prefix) {
		want, err := hex.DecodeString(strings.TrimPrefix(stored, otpHashV2Prefix))
		if err != nil {
			return false
		}
		h := hmac.New(sha256.New, otpHMACSecret())
		h.Write([]byte(plain))
		return hmac.Equal(want, h.Sum(nil))
	}
	return bcrypt.CompareHashAndPassword([]byte(stored), []byte(plain)) == nil
}

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

func (s *VerificationService) SendCode(email string, t verificationcodemodel.VerificationType) error {
	if err := verificationcodemodel.ValidateVerificationEmail(email); err != nil {
		return err
	}
	if err := verificationcodemodel.ValidateVerificationType(t); err != nil {
		return err
	}

	exists, err := s.userRepo.ExistsByEmail(email)
	if err != nil {
		return err
	}
	switch t {
	case verificationcodemodel.VerificationTypeRegister:
		if exists {
			return errors.New("email already registered")
		}
	case verificationcodemodel.VerificationTypeForgotPassword, verificationcodemodel.VerificationTypeLogin:
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
	record := &verificationcodemodel.VerificationCode{
		Email:     email,
		Code:      hashOTP(plain),
		Type:      t,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	if err := s.verificationRepo.Save(record); err != nil {
		return err
	}

	switch t {
	case verificationcodemodel.VerificationTypeRegister:
		return s.emailService.SendWelcomeOTP(email, plain)
	case verificationcodemodel.VerificationTypeForgotPassword:
		return s.emailService.SendPasswordResetOTP(email, plain)
	case verificationcodemodel.VerificationTypeLogin:
		return s.emailService.SendLoginOTP(email, plain)
	default:
		return errors.New("unsupported verification type")
	}
}

func (s *VerificationService) VerifyCode(email, code string, t verificationcodemodel.VerificationType) error {
	if err := verificationcodemodel.ValidateVerificationType(t); err != nil {
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

	if !verifyOTPHash(record.Code, code) {
		return errors.New("invalid code")
	}

	return s.verificationRepo.MarkUsed(record.ID)
}
