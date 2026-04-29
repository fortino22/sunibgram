package repository

import (
	"github.com/sunibgram/domain/model"
)

type IVerificationRepository interface {
	Save(code *model.VerificationCode) error
	FindLatestByEmailAndType(email string, t model.VerificationType) (*model.VerificationCode, error)
	FindVerifiedByEmailAndType(email string, t model.VerificationType) (*model.VerificationCode, error)
	MarkUsed(id uint) error
	DeleteByEmailAndType(email string, t model.VerificationType) error
}
