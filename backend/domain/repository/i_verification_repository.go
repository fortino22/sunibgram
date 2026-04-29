package repository

import "github.com/sunibgram/domain/model"

type IVerificationRepository interface {
	Save(code *model.VerificationCode) error
	FindLatestByEmail(email string) (*model.VerificationCode, error)
	MarkUsed(id uint) error
	DeleteByEmail(email string) error
}
