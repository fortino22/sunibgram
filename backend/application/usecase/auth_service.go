package usecase

import (
	"errors"
	"os"
	"time"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	accessTokenDuration  = 15 * time.Minute
	refreshTokenDuration = 7 * 24 * time.Hour
)

type tokenKind string

const (
	kindAccess  tokenKind = "access"
	kindRefresh tokenKind = "refresh"
)

type AuthService struct {
	userRepo         repository.IUserRepository
	roleRepo         repository.IRoleRepository
	verificationRepo repository.IVerificationRepository
}

func NewAuthService(userRepo repository.IUserRepository, roleRepo repository.IRoleRepository, verificationRepo repository.IVerificationRepository) service.IAuthService {
	return &AuthService{userRepo: userRepo, roleRepo: roleRepo, verificationRepo: verificationRepo}
}

func (s *AuthService) Register(user *model.User, confirmPassword string) error {
	if user.Password != confirmPassword {
		return errors.New("password and confirm password do not match")
	}
	if err := user.ValidateRegister(); err != nil {
		return err
	}

	if _, err := s.verificationRepo.FindVerifiedByEmailAndType(user.Email, model.VerificationTypeRegister); err != nil {
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

	defaultRole, err := s.roleRepo.FindByName(model.DefaultRoleName)
	if err != nil {
		return errors.New("default role not found, contact administrator")
	}
	user.RoleID = defaultRole.RoleID

	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashed)

	return s.userRepo.Save(user)
}

func (s *AuthService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("old password is incorrect")
	}

	if len(newPassword) < 6 {
		return errors.New("new password must be at least 6 characters")
	}

	if oldPassword == newPassword {
		return errors.New("new password must be different from old password")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePassword(userID, string(hashed))
}

func (s *AuthService) ForgotPassword(email, newPassword string) error {
	if _, err := s.verificationRepo.FindVerifiedByEmailAndType(email, model.VerificationTypeForgotPassword); err != nil {
		return errors.New("otp not verified or expired, please verify your otp first")
	}

	if len(newPassword) < 6 {
		return errors.New("new password must be at least 6 characters")
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

	_ = s.verificationRepo.DeleteByEmailAndType(email, model.VerificationTypeForgotPassword)
	return nil
}

func (s *AuthService) Login(email, password string) (model.TokenPair, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return model.TokenPair{}, errors.New("invalid email or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return model.TokenPair{}, errors.New("invalid email or password")
	}
	return generateTokenPair(user)
}

func (s *AuthService) LoginProd(email, password string) (*model.User, model.TokenPair, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, model.TokenPair{}, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, model.TokenPair{}, errors.New("invalid email or password")
	}

	if user.Role.Name != "admin" {
		return nil, model.TokenPair{}, errors.New("admin access required")
	}

	pair, err := generateTokenPair(user)
	if err != nil {
		return nil, model.TokenPair{}, err
	}

	return user, pair, nil
}

func (s *AuthService) RefreshToken(refreshToken string) (model.TokenPair, error) {
	claims, err := parseToken(refreshToken, kindRefresh)
	if err != nil {
		return model.TokenPair{}, errors.New("invalid or expired refresh token")
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return model.TokenPair{}, errors.New("invalid token claims")
	}

	user, err := s.userRepo.FindByID(uint(userID))
	if err != nil {
		return model.TokenPair{}, errors.New("user not found")
	}

	return generateTokenPair(user)
}

func generateTokenPair(user *model.User) (model.TokenPair, error) {
	secret := jwtSecret()
	now := time.Now()

	accessExp := now.Add(accessTokenDuration)
	accessToken, err := signToken(jwt.MapClaims{
		"kind":      string(kindAccess),
		"user_id":   user.UserID,
		"username":  user.UserName,
		"email":     user.Email,
		"role_id":   user.RoleID,
		"role_name": user.Role.Name,
		"exp":       accessExp.Unix(),
	}, secret)
	if err != nil {
		return model.TokenPair{}, err
	}

	refreshExp := now.Add(refreshTokenDuration)
	refreshToken, err := signToken(jwt.MapClaims{
		"kind":    string(kindRefresh),
		"user_id": user.UserID,
		"exp":     refreshExp.Unix(),
	}, secret)
	if err != nil {
		return model.TokenPair{}, err
	}

	return model.TokenPair{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  accessExp,
		RefreshTokenExpiresAt: refreshExp,
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
