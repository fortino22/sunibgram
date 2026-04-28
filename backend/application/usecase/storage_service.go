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
	"avatar":     "MINIO_BUCKET_AVATAR",
	"post":       "MINIO_BUCKET_POST",
	"document":   "MINIO_BUCKET_DOCUMENT",
	"assignment": "MINIO_BUCKET_ASSIGNMENT",
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

	return s.storage.Upload(ctx, bucket, objectName, file, header.Size, contentType)
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
