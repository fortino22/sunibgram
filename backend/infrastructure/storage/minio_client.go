package storage

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	domainservice "github.com/sunibgram/domain/service"
)

type MinioStorage struct {
	client       *minio.Client
	publicClient *minio.Client
	useSSL       bool
}

func NewMinioClient() domainservice.IStorageService {
	endpoint       := strings.TrimSpace(os.Getenv("MINIO_ENDPOINT"))
	accessKey      := strings.TrimSpace(os.Getenv("MINIO_ACCESS_KEY"))
	secretKey      := strings.TrimSpace(os.Getenv("MINIO_SECRET_KEY"))
	useSSL         := strings.TrimSpace(os.Getenv("MINIO_USE_SSL")) == "true"
	publicEndpoint := strings.TrimSpace(os.Getenv("MINIO_PUBLIC_ENDPOINT"))
	if publicEndpoint == "" {
		publicEndpoint = endpoint
	}

	transport := &http.Transport{
		Proxy:               func(*http.Request) (*url.URL, error) { return nil, nil },
		MaxIdleConns:        200,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  true,
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure:    useSSL,
		Transport: transport,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to connect to MinIO: %v", err))
	}

	publicClient, err := minio.New(publicEndpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure:    useSSL,
		Region:    "us-east-1",
		Transport: transport,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to create MinIO public client: %v", err))
	}

	return &MinioStorage{
		client:       client,
		publicClient: publicClient,
		useSSL:       useSSL,
	}
}

func bucketPolicyEnvKey(bucket string) string {
	key := strings.ToUpper(strings.ReplaceAll(bucket, "-", "_"))
	return "MINIO_POLICY_" + key
}

func (s *MinioStorage) ensureBucket(ctx context.Context, bucket string) error {
	exists, err := s.client.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket: %w", err)
	}
	if exists {
		return nil
	}

	if err := s.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
		if minio.ToErrorResponse(err).Code != "BucketAlreadyOwnedByYou" {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	policy := strings.ToLower(strings.TrimSpace(os.Getenv(bucketPolicyEnvKey(bucket))))
	if policy == "public" {
		publicPolicy := fmt.Sprintf(`{
			"Version":"2012-10-17",
			"Statement":[{
				"Effect":"Allow",
				"Principal":{"AWS":["*"]},
				"Action":["s3:GetObject"],
				"Resource":["arn:aws:s3:::%s/*"]
			}]
		}`, bucket)
		if err := s.client.SetBucketPolicy(ctx, bucket, publicPolicy); err != nil {
			return fmt.Errorf("failed to set bucket policy: %w", err)
		}
	}

	return nil
}

func (s *MinioStorage) Upload(ctx context.Context, bucket, objectName string, file multipart.File, size int64, contentType string) error {
	if err := s.ensureBucket(ctx, bucket); err != nil {
		return err
	}

	_, err := s.client.PutObject(ctx, bucket, objectName, file, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	return nil
}

func (s *MinioStorage) GetURL(bucket, objectName string) string {
	scheme := "http"
	if s.useSSL {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", scheme, s.publicClient.EndpointURL().Host, bucket, objectName)
}

func (s *MinioStorage) Delete(ctx context.Context, bucket, objectName string) error {
	if err := s.client.RemoveObject(ctx, bucket, objectName, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}
	return nil
}

func (s *MinioStorage) GeneratePresignedURL(_ context.Context, bucket, objectName string, expires time.Duration) (string, error) {
	u, err := s.publicClient.PresignedGetObject(context.Background(), bucket, objectName, expires, nil)
	if err != nil {
		return "", fmt.Errorf("presign failed: %w", err)
	}
	return u.String(), nil
}
