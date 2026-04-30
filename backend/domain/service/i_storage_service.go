package service

import (
	"context"
	"mime/multipart"
	"time"
)

type IStorageService interface {
	Upload(ctx context.Context, bucket, objectName string, file multipart.File, size int64, contentType string) error
	GetURL(bucket, objectName string) string
	Delete(ctx context.Context, bucket, objectName string) error
	GeneratePresignedURL(ctx context.Context, bucket, objectName string, expires time.Duration) (string, error)
}
