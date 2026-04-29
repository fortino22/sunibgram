package service

import "github.com/sunibgram/domain/model"

type IVerificationService interface {
	SendCode(email string, t model.VerificationType) error
	VerifyCode(email, code string, t model.VerificationType) error
}
