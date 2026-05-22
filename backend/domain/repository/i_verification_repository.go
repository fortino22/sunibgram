package repository

import (
	verificationcodemodel "github.com/sunibgram/domain/model/verification_code"
)

type IVerificationRepository interface {
	Save(code *verificationcodemodel.VerificationCode) error
	FindLatestByEmailAndType(email string, t verificationcodemodel.VerificationType) (*verificationcodemodel.VerificationCode, error)
	FindVerifiedByEmailAndType(email string, t verificationcodemodel.VerificationType) (*verificationcodemodel.VerificationCode, error)
	MarkUsed(id uint) error
	DeleteByEmailAndType(email string, t verificationcodemodel.VerificationType) error
}
