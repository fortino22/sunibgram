package storageuc

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	storagesvciface "github.com/sunibgram/domain/service/storage"
)

// Two-bucket S3 model. Bucket name = "media" (public, CloudFront) or "private"
// (presigned). Object key = "<prefix>/<uuid>.<ext>". Both bucket+prefix are
// derived from upload type; bucket no longer leaks into the DB.
type bucketPrefix struct {
	Bucket string
	Prefix string
}

const (
	bucketMedia   = "media"
	bucketPrivate = "private"
)

var bucketPrefixMap = map[string]bucketPrefix{
	"avatar":           {bucketMedia, "avatar"},
	"banner":           {bucketMedia, "banner"},
	"post":             {bucketMedia, "post"},
	"poll":             {bucketMedia, "poll"},
	"community_banner": {bucketMedia, "community-banner"},
	"event_banner":     {bucketMedia, "event-banner"},
	"group_avatar":     {bucketMedia, "group-avatar"},
	"chat":             {bucketPrivate, "chat"},
	"document_event":   {bucketPrivate, "document-event"},
}

type StorageUsecase struct {
	storage storagesvciface.IStorageService
}

func NewStorageUsecase(storage storagesvciface.IStorageService) *StorageUsecase {
	return &StorageUsecase{storage: storage}
}

// ResolveBucketAndPrefix returns the logical bucket name ("media" or "private")
// and the S3 key prefix for the given upload type.
func (s *StorageUsecase) ResolveBucketAndPrefix(uploadType string) (string, string, error) {
	bp, ok := bucketPrefixMap[strings.ToLower(uploadType)]
	if !ok {
		return "", "", fmt.Errorf("upload type '%s' is not allowed", uploadType)
	}
	return bp.Bucket, bp.Prefix, nil
}

func (s *StorageUsecase) Upload(ctx context.Context, prefix string, file multipart.File, header *multipart.FileHeader) (string, error) {
	ext := filepath.Ext(header.Filename)
	objectName := fmt.Sprintf("%s/%s%s", prefix, uuid.New().String(), ext)

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	if err := s.storage.Upload(ctx, objectName, file, header.Size, contentType); err != nil {
		return "", err
	}
	return objectName, nil
}

func (s *StorageUsecase) GetURL(objectName string) string {
	return s.storage.GetURL(objectName)
}

func (s *StorageUsecase) Delete(ctx context.Context, objectName string) error {
	return s.storage.Delete(ctx, objectName)
}

func (s *StorageUsecase) GetAllowedTypes() []string {
	types := make([]string, 0, len(bucketPrefixMap))
	for k := range bucketPrefixMap {
		types = append(types, k)
	}
	return types
}

func (s *StorageUsecase) GeneratePresignedURL(ctx context.Context, objectName string) (string, error) {
	return s.storage.GeneratePresignedURL(ctx, objectName, 24*time.Hour)
}
