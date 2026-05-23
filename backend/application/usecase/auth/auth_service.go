package authuc

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	rolemodel "github.com/sunibgram/domain/model/role"
	usermodel "github.com/sunibgram/domain/model/user"
	userprofilemodel "github.com/sunibgram/domain/model/user_profile"
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
	cache            *cache.Cache
	otpFailLimiter   *ratelimit.RateLimiter
}

func NewAuthService(
	userRepo userrepoiface.IUserRepository,
	roleRepo rolerepoiface.IRoleRepository,
	verificationRepo verificationrepoiface.IVerificationRepository,
	userSettingsRepo usersettingsrepoiface.IUserSettingsRepository,
	c *cache.Cache,
) authsvciface.IAuthService {
	return &AuthService{
		userRepo:         userRepo,
		roleRepo:         roleRepo,
		verificationRepo: verificationRepo,
		userSettingsRepo: userSettingsRepo,
		cache:            c,
		otpFailLimiter:   ratelimit.New(c, "auth_otp_fail", 5, 10*time.Minute),
	}
}

func sessionKey(sid string) string {
	return fmt.Sprintf("session:%s", sid)
}

func userSessionsKey(userID uint) string {
	return fmt.Sprintf("user_sessions:%d", userID)
}

func (s *AuthService) Register(user *usermodel.User, profile *userprofilemodel.UserProfile, confirmPassword string) error {
	if user.Password != confirmPassword {
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

	if err := s.userRepo.SaveWithProfile(user, profile); err != nil {
		return err
	}
	return s.userSettingsRepo.Create(context.Background(), &usersettingsmodel.UserSettings{UserID: user.UserID})
}

func (s *AuthService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("old password is incorrect")
	}

	if err := usermodel.ValidatePassword(newPassword); err != nil {
		return err
	}

	if oldPassword == newPassword {
		return errors.New("new password must be different from old password")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := s.userRepo.UpdatePassword(userID, string(hashed)); err != nil {
		return err
	}
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
	if err := bcrypt.CompareHashAndPassword([]byte(record.Code), []byte(otp)); err != nil {
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

	if user.Role.Name != "admin" {
		return nil, usermodel.TokenPair{}, errors.New("admin access required")
	}

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

	var userID uint
	if err := s.cache.Get(ctx, sessionKey(sid), &userID); err != nil {
		return usermodel.TokenPair{}, errors.New("session expired or revoked")
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return usermodel.TokenPair{}, errors.New("user not found")
	}

	dur := sessionDuration()
	if err := s.cache.Set(ctx, sessionKey(sid), userID, dur); err != nil {
		return usermodel.TokenPair{}, errors.New("failed to extend session")
	}
	_ = s.cache.SAdd(ctx, userSessionsKey(userID), sid, dur)

	return generateTokenPair(user, sid)
}

func (s *AuthService) Logout(ctx context.Context, sid string) error {
	if sid == "" {
		return errors.New("invalid session")
	}
	var userID uint
	if err := s.cache.Get(ctx, sessionKey(sid), &userID); err == nil {
		_ = s.cache.SRem(ctx, userSessionsKey(userID), sid)
	}
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
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
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
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return s
	}
	return "sunibgram-secret"
}
