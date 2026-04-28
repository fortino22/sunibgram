package storage

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"time"
	"strings"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	domainservice "github.com/sunibgram/domain/service"
)

type MinioStorage struct {
	client   *minio.Client
	endpoint string
	useSSL   bool
}

func NewMinioClient() domainservice.IStorageService {
	endpoint  := strings.TrimSpace(os.Getenv("MINIO_ENDPOINT"))
	accessKey := strings.TrimSpace(os.Getenv("MINIO_ACCESS_KEY"))
	secretKey := strings.TrimSpace(os.Getenv("MINIO_SECRET_KEY"))

	useSSL := strings.TrimSpace(os.Getenv("MINIO_USE_SSL")) == "true"

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})

	if err != nil {
		panic(fmt.Sprintf("failed to connect to MinIO: %v", err))
	}

	return &MinioStorage{client: client, endpoint: endpoint, useSSL: useSSL}
}

func (s *MinioStorage) ensureBucket(ctx context.Context, bucket string) error {
	exists, err := s.client.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket: %w", err)
	}
	if !exists {
		if err := s.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}
	return nil
}

func (s *MinioStorage) Upload(ctx context.Context, bucket, objectName string, file multipart.File, size int64, contentType string) (string, error) {
	if err := s.ensureBucket(ctx, bucket); err != nil {
		return "", err
	}

	_, err := s.client.PutObject(ctx, bucket, objectName, file, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("upload failed: %w", err)
	}

	scheme := "http"
	if s.useSSL {
		scheme = "https"
	}

	url := fmt.Sprintf("%s://%s/%s/%s", scheme, s.endpoint, bucket, objectName)
	return url, nil
}

func (s *MinioStorage) GeneratePresignedURL(ctx context.Context, bucket, objectName string, expires time.Duration) (string, error) {
	url, err := s.client.PresignedGetObject(ctx, bucket, objectName, expires, nil)
	if err != nil {
		return "", fmt.Errorf("presign failed: %w", err)
	}
	return url.String(), nil
}

