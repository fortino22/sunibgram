package usecase

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sunibgram/domain/service"
)

var bucketWhitelist = map[string]string{
	"avatar":           "MINIO_BUCKET_AVATAR",
	"banner":           "MINIO_BUCKET_BANNER",
	"post":             "MINIO_BUCKET_POST",
	"chat":             "MINIO_BUCKET_CHAT",
	"community_avatar": "MINIO_BUCKET_COMMUNITY_AVATAR",
	"community_banner": "MINIO_BUCKET_COMMUNITY_BANNER",
	"event_banner":     "MINIO_BUCKET_EVENT_BANNER",
	"document_event":   "MINIO_BUCKET_DOCUMENT_EVENT",
}

type StorageUsecase struct {
	storage service.IStorageService
}

func NewStorageUsecase(storage service.IStorageService) *StorageUsecase {
	return &StorageUsecase{storage: storage}
}

func (s *StorageUsecase) ResolveBucket(uploadType string) (string, error) {
	envKey, allowed := bucketWhitelist[strings.ToLower(uploadType)]
	if !allowed {
		return "", fmt.Errorf("upload type '%s' is not allowed", uploadType)
	}

	bucket := os.Getenv(envKey)
	if bucket == "" {
		return "", fmt.Errorf("bucket for type '%s' is not configured", uploadType)
	}

	return bucket, nil
}

func (s *StorageUsecase) Upload(ctx context.Context, bucket string, file multipart.File, header *multipart.FileHeader) (string, error) {
	ext := filepath.Ext(header.Filename)
	objectName := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	if err := s.storage.Upload(ctx, bucket, objectName, file, header.Size, contentType); err != nil {
		return "", err
	}
	return objectName, nil
}

func (s *StorageUsecase) GetURL(bucket, objectName string) string {
	return s.storage.GetURL(bucket, objectName)
}

func (s *StorageUsecase) Delete(ctx context.Context, bucket, objectName string) error {
	return s.storage.Delete(ctx, bucket, objectName)
}

func (s *StorageUsecase) GetAllowedTypes() []string {
	types := make([]string, 0, len(bucketWhitelist))
	for k := range bucketWhitelist {
		types = append(types, k)
	}
	return types
}

func (s *StorageUsecase) GeneratePresignedURL(ctx context.Context, bucket, objectName string) (string, error) {
	return s.storage.GeneratePresignedURL(ctx, bucket, objectName, 24*time.Hour)
}
