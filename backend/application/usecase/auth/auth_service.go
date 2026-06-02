package authuc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/sunibgram/application/helper"
	verificationuc "github.com/sunibgram/application/usecase/verification"
	storageuc "github.com/sunibgram/application/usecase/storage"
	mediamodel "github.com/sunibgram/domain/model/media"
	rolemodel "github.com/sunibgram/domain/model/role"
	usermodel "github.com/sunibgram/domain/model/user"
	usersettingsmodel "github.com/sunibgram/domain/model/user_settings"
	verificationcodemodel "github.com/sunibgram/domain/model/verification_code"
	rolerepoiface "github.com/sunibgram/domain/repository/role"
	userrepoiface "github.com/sunibgram/domain/repository/user"
	usersettingsrepoiface "github.com/sunibgram/domain/repository/user_settings"
	verificationrepoiface "github.com/sunibgram/domain/repository/verification"
	authsvciface "github.com/sunibgram/domain/service/auth"
	"github.com/sunibgram/infrastructure/cache"
	"github.com/sunibgram/infrastructure/ratelimit"
)

func accessTokenDuration() time.Duration {
	if v := os.Getenv("JWT_ACCESS_TOKEN_DURATION"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return time.Hour
}

func sessionDuration() time.Duration {
	if v := os.Getenv("JWT_SESSION_DURATION"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return 7 * 24 * time.Hour
}

type tokenKind string

const kindAccess tokenKind = "access"

type AuthService struct {
	userRepo         userrepoiface.IUserRepository
	roleRepo         rolerepoiface.IRoleRepository
	verificationRepo verificationrepoiface.IVerificationRepository
	userSettingsRepo usersettingsrepoiface.IUserSettingsRepository
	storageUsecase   *storageuc.StorageUsecase
	cache            *cache.Cache
	otpFailLimiter   *ratelimit.RateLimiter
	refreshLimiter   *ratelimit.RateLimiter
}

func NewAuthService(
	userRepo userrepoiface.IUserRepository,
	roleRepo rolerepoiface.IRoleRepository,
	verificationRepo verificationrepoiface.IVerificationRepository,
	userSettingsRepo usersettingsrepoiface.IUserSettingsRepository,
	storageUsecase *storageuc.StorageUsecase,
	c *cache.Cache,
) authsvciface.IAuthService {
	return &AuthService{
		userRepo:         userRepo,
		roleRepo:         roleRepo,
		verificationRepo: verificationRepo,
		userSettingsRepo: userSettingsRepo,
		storageUsecase:   storageUsecase,
		cache:            c,
		otpFailLimiter:   ratelimit.New(c, "auth_otp_fail", 5, 10*time.Minute),
		refreshLimiter:   ratelimit.New(c, "auth_refresh", 60, time.Minute),
	}
}

const refreshResultTTL = 3 * time.Second

func refreshResultKey(sid string) string {
	return fmt.Sprintf("refresh:result:%s", sid)
}

func sessionKey(sid string) string {
	return fmt.Sprintf("session:%s", sid)
}

func userSessionsKey(userID uint) string {
	return fmt.Sprintf("user_sessions:%d", userID)
}

// IsUsernameAvailable reports whether a username is free. Public (used by the
// signup form before the user has a token); reuses the same uniqueness check
// Register relies on so the two never disagree.
func (s *AuthService) IsUsernameAvailable(username string) (bool, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return false, errors.New("username is required")
	}
	exists, err := s.userRepo.ExistsByUserName(username)
	if err != nil {
		return false, err
	}
	return !exists, nil
}

// Register creates an account atomically. All fallible external work (image
// validation + S3 upload) happens before any DB write, and the user, profile,
// settings, and faculty assignment are persisted in a single transaction
// (userRepo.RegisterAccount). A failure at any step leaves nothing behind, so a
// retry with the same email is not blocked by a half-created orphan row.
func (s *AuthService) Register(ctx context.Context, input authsvciface.RegisterInput) error {
	user := input.User
	profile := input.Profile

	if user.Password != input.ConfirmPassword {
		return errors.New("password and confirm password do not match")
	}
	if err := user.ValidateRegister(); err != nil {
		return err
	}
	if profile == nil {
		return errors.New("profile is required")
	}
	if err := profile.ValidateRequiredFields(); err != nil {
		return err
	}

	if _, err := s.verificationRepo.FindVerifiedByEmailAndType(user.Email, verificationcodemodel.VerificationTypeRegister); err != nil {
		return errors.New("email not verified, please verify your otp first")
	}

	// Fail fast on bad image types before any side effect (S3 upload / DB write).
	if input.Avatar != nil {
		if err := mediamodel.ValidateImageExt(input.Avatar.Filename); err != nil {
			return err
		}
	}
	if input.Banner != nil {
		if err := mediamodel.ValidateImageExt(input.Banner.Filename); err != nil {
			return err
		}
	}

	emailExists, err := s.userRepo.ExistsByEmail(user.Email)
	if err != nil {
		return err
	}
	if emailExists {
		return errors.New("email already registered")
	}

	usernameExists, err := s.userRepo.ExistsByUserName(user.UserName)
	if err != nil {
		return err
	}
	if usernameExists {
		return errors.New("username already taken")
	}

	defaultRole, err := s.roleRepo.FindByName(rolemodel.DefaultRoleName)
	if err != nil {
		return errors.New("default role not found, contact administrator")
	}
	user.RoleID = defaultRole.RoleID

	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashed)

	// Upload media before any DB write so a failed/slow upload never orphans an
	// account. helper.UploadImage re-validates size/ext/content.
	if input.Avatar != nil {
		objectName, err := helper.UploadImage(ctx, s.storageUsecase, input.Avatar, "avatar")
		if err != nil {
			return err
		}
		profile.AvatarObjectName = objectName
	}
	if input.Banner != nil {
		objectName, err := helper.UploadImage(ctx, s.storageUsecase, input.Banner, "banner")
		if err != nil {
			return err
		}
		profile.BannerObjectName = objectName
	}
	if bio := strings.TrimSpace(input.Bio); bio != "" {
		if utf8.RuneCountInString(bio) > 500 {
			return errors.New("bio must be 500 characters or less")
		}
		profile.Bio = bio
	}

	if err := s.userRepo.RegisterAccount(user, profile, &usersettingsmodel.UserSettings{}, input.FacultyID); err != nil {
		// No row landed — best-effort delete the just-uploaded objects so a failed
		// registration doesn't leak orphaned media in the bucket.
		if profile.AvatarObjectName != "" {
			_ = s.storageUsecase.Delete(ctx, profile.AvatarObjectName)
		}
		if profile.BannerObjectName != "" {
			_ = s.storageUsecase.Delete(ctx, profile.BannerObjectName)
		}
		return err
	}
	return nil
}

// ChangePassword changes a logged-in user's password after an email OTP of type
// CHANGE_PASSWORD has been verified server-side. Identity comes from the JWT
// (userID); email is the account email from claims and must match the account.
func (s *AuthService) ChangePassword(userID uint, email, newPassword string) error {
	if _, err := s.verificationRepo.FindVerifiedByEmailAndType(email, verificationcodemodel.VerificationTypeChangePassword); err != nil {
		return errors.New("otp not verified or expired, please verify your otp first")
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}
	if user.Email != email {
		return errors.New("email mismatch")
	}

	if err := usermodel.ValidatePassword(newPassword); err != nil {
		return err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := s.userRepo.UpdatePassword(userID, string(hashed)); err != nil {
		return err
	}

	_ = s.verificationRepo.DeleteByEmailAndType(email, verificationcodemodel.VerificationTypeChangePassword)
	s.revokeAllSessions(context.Background(), userID)
	return nil
}

func (s *AuthService) ForgotPassword(email, newPassword string) error {
	if _, err := s.verificationRepo.FindVerifiedByEmailAndType(email, verificationcodemodel.VerificationTypeForgotPassword); err != nil {
		return errors.New("otp not verified or expired, please verify your otp first")
	}

	if err := usermodel.ValidatePassword(newPassword); err != nil {
		return err
	}

	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return errors.New("user not found")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := s.userRepo.UpdatePassword(user.UserID, string(hashed)); err != nil {
		return err
	}

	_ = s.verificationRepo.DeleteByEmailAndType(email, verificationcodemodel.VerificationTypeForgotPassword)
	s.revokeAllSessions(context.Background(), user.UserID)
	return nil
}

func (s *AuthService) revokeAllSessions(ctx context.Context, userID uint) {
	sids, err := s.cache.SMembers(ctx, userSessionsKey(userID))
	if err != nil {
		return
	}
	keys := make([]string, len(sids))
	for i, sid := range sids {
		keys[i] = sessionKey(sid)
	}
	if len(keys) > 0 {
		_ = s.cache.Delete(ctx, keys...)
	}
	_ = s.cache.Delete(ctx, userSessionsKey(userID))
}

// VerifyCredentials checks email + password without issuing a session. Used by
// the login-OTP send path so a wrong password is rejected up front instead of
// emailing a code and only failing at the final /login step.
func (s *AuthService) VerifyCredentials(email, password string) error {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return errors.New("invalid email or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return errors.New("invalid email or password")
	}
	return nil
}

func (s *AuthService) LoginWithOTP(email, password, otp string) (usermodel.TokenPair, error) {
	if err := s.otpFailLimiter.Check(email); err != nil {
		return usermodel.TokenPair{}, err
	}

	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return usermodel.TokenPair{}, errors.New("invalid email or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return usermodel.TokenPair{}, errors.New("invalid email or password")
	}

	record, err := s.verificationRepo.FindLatestByEmailAndType(email, verificationcodemodel.VerificationTypeLogin)
	if err != nil {
		s.otpFailLimiter.Fail(email)
		return usermodel.TokenPair{}, errors.New("otp not found, please request a new one")
	}
	if time.Now().After(record.ExpiresAt) {
		s.otpFailLimiter.Fail(email)
		return usermodel.TokenPair{}, errors.New("otp has expired")
	}
	if !verificationuc.VerifyOTPHash(record.Code, otp) {
		s.otpFailLimiter.Fail(email)
		return usermodel.TokenPair{}, errors.New("invalid otp")
	}
	_ = s.verificationRepo.DeleteByEmailAndType(email, verificationcodemodel.VerificationTypeLogin)

	sid := uuid.NewString()
	ctx := context.Background()
	dur := sessionDuration()
	if err := s.cache.Set(ctx, sessionKey(sid), user.UserID, dur); err != nil {
		return usermodel.TokenPair{}, errors.New("failed to create session")
	}
	_ = s.cache.SAdd(ctx, userSessionsKey(user.UserID), sid, dur)

	return generateTokenPair(user, sid)
}

func (s *AuthService) Login(email, password string) (usermodel.TokenPair, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return usermodel.TokenPair{}, errors.New("invalid email or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return usermodel.TokenPair{}, errors.New("invalid email or password")
	}

	sid := uuid.NewString()
	ctx := context.Background()
	dur := sessionDuration()
	if err := s.cache.Set(ctx, sessionKey(sid), user.UserID, dur); err != nil {
		return usermodel.TokenPair{}, errors.New("failed to create session")
	}
	_ = s.cache.SAdd(ctx, userSessionsKey(user.UserID), sid, dur)

	return generateTokenPair(user, sid)
}

func (s *AuthService) LoginProd(email, password string) (*usermodel.User, usermodel.TokenPair, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, usermodel.TokenPair{}, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, usermodel.TokenPair{}, errors.New("invalid email or password")
	}

	// if user.Role.Name != "admin" {
	// 	return nil, usermodel.TokenPair{}, errors.New("admin access required")
	// }

	sid := uuid.NewString()
	ctx := context.Background()
	dur := sessionDuration()
	if err := s.cache.Set(ctx, sessionKey(sid), user.UserID, dur); err != nil {
		return nil, usermodel.TokenPair{}, errors.New("failed to create session")
	}
	_ = s.cache.SAdd(ctx, userSessionsKey(user.UserID), sid, dur)

	pair, err := generateTokenPair(user, sid)
	if err != nil {
		return nil, usermodel.TokenPair{}, err
	}

	return user, pair, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, sid string) (usermodel.TokenPair, error) {
	if sid == "" {
		return usermodel.TokenPair{}, errors.New("sid is required")
	}

	// Idempotent response cache — spam within window returns same pair.
	var cached usermodel.TokenPair
	if err := s.cache.Get(ctx, refreshResultKey(sid), &cached); err == nil {
		return cached, nil
	}

	if err := s.refreshLimiter.Allow(sid); err != nil {
		return usermodel.TokenPair{}, authsvciface.ErrRefreshRateLimited
	}

	var userID uint
	if err := s.cache.Get(ctx, sessionKey(sid), &userID); err != nil {
		if s.cache.Miss(err) {
			return usermodel.TokenPair{}, authsvciface.ErrSessionRevoked
		}
		return usermodel.TokenPair{}, authsvciface.ErrSessionTransient
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return usermodel.TokenPair{}, errors.New("user not found")
	}

	dur := sessionDuration()
	if err := s.cache.Set(ctx, sessionKey(sid), userID, dur); err != nil {
		return usermodel.TokenPair{}, authsvciface.ErrSessionTransient
	}
	_ = s.cache.SAdd(ctx, userSessionsKey(userID), sid, dur)

	pair, err := generateTokenPair(user, sid)
	if err != nil {
		return usermodel.TokenPair{}, err
	}

	_ = s.cache.Set(ctx, refreshResultKey(sid), pair, refreshResultTTL)
	return pair, nil
}

func (s *AuthService) Logout(ctx context.Context, sid string) error {
	if sid == "" {
		return errors.New("invalid session")
	}
	var userID uint
	if err := s.cache.Get(ctx, sessionKey(sid), &userID); err == nil {
		_ = s.cache.SRem(ctx, userSessionsKey(userID), sid)
	}
	_ = s.cache.Delete(ctx, refreshResultKey(sid))
	return s.cache.Delete(ctx, sessionKey(sid))
}

func generateTokenPair(user *usermodel.User, sid string) (usermodel.TokenPair, error) {
	secret := jwtSecret()
	now := time.Now()
	accessExp := now.Add(accessTokenDuration())

	accessToken, err := signToken(jwt.MapClaims{
		"kind":      string(kindAccess),
		"sid":       sid,
		"user_id":   user.UserID,
		"username":  user.UserName,
		"email":     user.Email,
		"role_id":   user.RoleID,
		"role_name": user.Role.Name,
		"exp":       accessExp.Unix(),
	}, secret)
	if err != nil {
		return usermodel.TokenPair{}, err
	}

	return usermodel.TokenPair{
		AccessToken:          accessToken,
		SID:                  sid,
		AccessTokenExpiresAt: accessExp,
	}, nil
}

func parseToken(tokenStr string, expected tokenKind) (jwt.MapClaims, error) {
	secret := jwtSecret()
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	if kind, _ := claims["kind"].(string); tokenKind(kind) != expected {
		return nil, errors.New("wrong token type")
	}

	return claims, nil
}

func signToken(claims jwt.MapClaims, secret string) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

func jwtSecret() string {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		log.Fatal("JWT_SECRET is required")
	}
	return s
}
