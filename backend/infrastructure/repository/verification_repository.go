package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
)

type VerificationRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewVerificationRepository(writeDB, readDB *gorm.DB) domainrepo.IVerificationRepository {
	return &VerificationRepository{writeDB: writeDB, readDB: readDB}
}

func (r *VerificationRepository) Save(code *model.VerificationCode) error {
	return r.writeDB.Create(code).Error
}

func (r *VerificationRepository) FindLatestByEmailAndType(email string, t model.VerificationType) (*model.VerificationCode, error) {
	var code model.VerificationCode
	err := r.readDB.Where("email = ? AND type = ? AND used = false", email, t).
		Order("created_at DESC").
		First(&code).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("code not found")
	}
	return &code, err
}

func (r *VerificationRepository) FindVerifiedByEmailAndType(email string, t model.VerificationType) (*model.VerificationCode, error) {
	var code model.VerificationCode
	err := r.readDB.Where("email = ? AND type = ? AND used = true AND expires_at > ?", email, t, time.Now()).
		Order("created_at DESC").
		First(&code).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("otp not verified")
	}
	return &code, err
}

func (r *VerificationRepository) MarkUsed(id uint) error {
	return r.writeDB.Model(&model.VerificationCode{}).Where("id = ?", id).Update("used", true).Error
}

func (r *VerificationRepository) DeleteByEmailAndType(email string, t model.VerificationType) error {
	return r.writeDB.Where("email = ? AND type = ?", email, t).Delete(&model.VerificationCode{}).Error
}
