package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"

	verificationcodemodel "github.com/sunibgram/domain/model/verification_code"
	verificationrepoiface "github.com/sunibgram/domain/repository/verification"
)

type VerificationRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewVerificationRepository(writeDB, readDB *gorm.DB) verificationrepoiface.IVerificationRepository {
	return &VerificationRepository{writeDB: writeDB, readDB: readDB}
}

func (r *VerificationRepository) Save(code *verificationcodemodel.VerificationCode) error {
	return r.writeDB.Create(code).Error
}

func (r *VerificationRepository) FindLatestByEmailAndType(email string, t verificationcodemodel.VerificationType) (*verificationcodemodel.VerificationCode, error) {
	var code verificationcodemodel.VerificationCode
	err := r.readDB.Where("email = ? AND type = ? AND used = false", email, t).
		Order("created_at DESC").
		First(&code).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("code not found")
	}
	return &code, err
}

func (r *VerificationRepository) FindVerifiedByEmailAndType(email string, t verificationcodemodel.VerificationType) (*verificationcodemodel.VerificationCode, error) {
	var code verificationcodemodel.VerificationCode
	err := r.readDB.Where("email = ? AND type = ? AND used = true AND expires_at > ?", email, t, time.Now()).
		Order("created_at DESC").
		First(&code).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("otp not verified")
	}
	return &code, err
}

func (r *VerificationRepository) MarkUsed(id uint) error {
	return r.writeDB.Model(&verificationcodemodel.VerificationCode{}).Where("id = ?", id).Update("used", true).Error
}

func (r *VerificationRepository) DeleteByEmailAndType(email string, t verificationcodemodel.VerificationType) error {
	return r.writeDB.Where("email = ? AND type = ?", email, t).Delete(&verificationcodemodel.VerificationCode{}).Error
}
