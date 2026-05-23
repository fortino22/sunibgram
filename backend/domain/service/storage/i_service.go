package storagesvciface

import (
	"context"
	"mime/multipart"
	"time"
)

type IStorageService interface {
	Upload(ctx context.Context, objectName string, file multipart.File, size int64, contentType string) error
	GetURL(objectName string) string
	Delete(ctx context.Context, objectName string) error
	GeneratePresignedURL(ctx context.Context, objectName string, expires time.Duration) (string, error)
}
