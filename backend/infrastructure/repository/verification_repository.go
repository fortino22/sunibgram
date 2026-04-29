package repository

import (
	"errors"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"gorm.io/gorm"
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

func (r *VerificationRepository) FindLatestByEmail(email string) (*model.VerificationCode, error) {
	var code model.VerificationCode
	err := r.readDB.Where("email = ? AND used = false", email).
		Order("created_at DESC").
		First(&code).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("code not found")
	}
	return &code, err
}

func (r *VerificationRepository) MarkUsed(id uint) error {
	return r.writeDB.Model(&model.VerificationCode{}).Where("id = ?", id).Update("used", true).Error
}

func (r *VerificationRepository) DeleteByEmail(email string) error {
	return r.writeDB.Where("email = ?", email).Delete(&model.VerificationCode{}).Error
}
