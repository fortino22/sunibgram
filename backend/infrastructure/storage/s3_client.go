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

	storagesvciface "github.com/sunibgram/domain/service/storage"
)

type S3Storage struct {
	client             *minio.Client
	publicClient       *minio.Client
	useSSL             bool
	cdnBaseURL         string
	autoCreateBuckets  bool
}

func NewS3Client() storagesvciface.IStorageService {
	endpoint := strings.TrimSpace(os.Getenv("S3_ENDPOINT"))
	accessKey := strings.TrimSpace(os.Getenv("S3_ACCESS_KEY"))
	secretKey := strings.TrimSpace(os.Getenv("S3_SECRET_KEY"))
	useSSL := strings.EqualFold(strings.TrimSpace(os.Getenv("S3_USE_SSL")), "true")
	publicEndpoint := strings.TrimSpace(os.Getenv("S3_PUBLIC_ENDPOINT"))
	if publicEndpoint == "" {
		publicEndpoint = endpoint
	}
	region := strings.TrimSpace(os.Getenv("S3_REGION"))
	cdnBaseURL := strings.TrimSpace(os.Getenv("S3_CDN_BASE_URL"))
	useIAMRole := strings.EqualFold(strings.TrimSpace(os.Getenv("S3_USE_IAM_ROLE")), "true")
	autoCreateBuckets := strings.EqualFold(strings.TrimSpace(os.Getenv("S3_AUTO_CREATE_BUCKETS")), "true")

	transport := &http.Transport{
		Proxy:               func(*http.Request) (*url.URL, error) { return nil, nil },
		MaxIdleConns:        200,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  true,
	}

	var creds *credentials.Credentials
	if useIAMRole {
		creds = credentials.NewIAM("")
	} else {
		creds = credentials.NewStaticV4(accessKey, secretKey, "")
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:     creds,
		Secure:    useSSL,
		Region:    region,
		Transport: transport,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to connect to S3: %v", err))
	}

	publicClient, err := minio.New(publicEndpoint, &minio.Options{
		Creds:     creds,
		Secure:    useSSL,
		Region:    region,
		Transport: transport,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to create S3 public client: %v", err))
	}

	return &S3Storage{
		client:            client,
		publicClient:      publicClient,
		useSSL:            useSSL,
		cdnBaseURL:        cdnBaseURL,
		autoCreateBuckets: autoCreateBuckets,
	}
}

func bucketPolicyEnvValue(bucket string) string {
	key := strings.ToUpper(strings.ReplaceAll(bucket, "-", "_"))
	return strings.TrimSpace(os.Getenv("S3_POLICY_" + key))
}

func (s *S3Storage) ensureBucket(ctx context.Context, bucket string) error {
	if !s.autoCreateBuckets {
		return nil
	}

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

	policy := strings.ToLower(bucketPolicyEnvValue(bucket))
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

func (s *S3Storage) Upload(ctx context.Context, bucket, objectName string, file multipart.File, size int64, contentType string) error {
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

func (s *S3Storage) GetURL(bucket, objectName string) string {
	if s.cdnBaseURL != "" {
		return fmt.Sprintf("%s/%s/%s", strings.TrimRight(s.cdnBaseURL, "/"), bucket, objectName)
	}
	scheme := "http"
	if s.useSSL {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", scheme, s.publicClient.EndpointURL().Host, bucket, objectName)
}

func (s *S3Storage) Delete(ctx context.Context, bucket, objectName string) error {
	if err := s.client.RemoveObject(ctx, bucket, objectName, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}
	return nil
}

func (s *S3Storage) GeneratePresignedURL(_ context.Context, bucket, objectName string, expires time.Duration) (string, error) {
	u, err := s.publicClient.PresignedGetObject(context.Background(), bucket, objectName, expires, nil)
	if err != nil {
		return "", fmt.Errorf("presign failed: %w", err)
	}
	return u.String(), nil
}
