package usecase

import (
	"context"
	"errors"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

var allowedAvatarExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".webp": true,
}

type AvatarService struct {
	avatarRepo     repository.IUserAvatarRepository
	storageUsecase *StorageUsecase
}

func NewAvatarService(avatarRepo repository.IUserAvatarRepository, storageUsecase *StorageUsecase) service.IAvatarService {
	return &AvatarService{avatarRepo: avatarRepo, storageUsecase: storageUsecase}
}

func (s *AvatarService) UploadAvatar(ctx context.Context, userID uint, file multipart.File, header *multipart.FileHeader) (*model.UserAvatar, error) {
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedAvatarExts[ext] {
		return nil, errors.New("avatar must be jpg, jpeg, png, or webp")
	}

	bucket, err := s.storageUsecase.ResolveBucket("avatar")
	if err != nil {
		return nil, err
	}

	url, err := s.storageUsecase.Upload(ctx, bucket, file, header)
	if err != nil {
		return nil, err
	}

	objectName := filepath.Base(url)

	exists, err := s.avatarRepo.ExistsByUserID(userID)
	if err != nil {
		return nil, err
	}

	if exists {
		existing, err := s.avatarRepo.FindByUserID(userID)
		if err != nil {
			return nil, err
		}
		existing.URL = url
		existing.ObjectName = objectName
		existing.Bucket = bucket
		if err := s.avatarRepo.Update(existing); err != nil {
			return nil, err
		}
		return existing, nil
	}

	avatar := &model.UserAvatar{
		UserID:     userID,
		ObjectName: objectName,
		Bucket:     bucket,
		URL:        url,
	}

	if err := s.avatarRepo.Save(avatar); err != nil {
		return nil, err
	}

	return avatar, nil
}

func (s *AvatarService) GetAvatarByUserID(userID uint) (*model.UserAvatar, error) {
	return s.avatarRepo.FindByUserID(userID)
}

func (s *AvatarService) DeleteAvatarByUserID(userID uint) error {
	exists, err := s.avatarRepo.ExistsByUserID(userID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("avatar not found")
	}
	return s.avatarRepo.DeleteByUserID(userID)
}
