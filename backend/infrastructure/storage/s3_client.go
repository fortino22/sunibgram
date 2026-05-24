package storage

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	storagesvciface "github.com/sunibgram/domain/service/storage"
)

type presignedEntry struct {
	url     string
	expires time.Time
}

const presignSafetyMargin = 5 * time.Minute

// S3Storage talks to AWS S3 via aws-sdk-go-v2. Credentials are resolved via the
// SDK default chain (env vars → ~/.aws/credentials with SSO profile → IAM role).
// Two buckets: media (public, CloudFront-fronted) and private (presigned only).
// The bucket for any operation is derived from the object key's top-level prefix.
type S3Storage struct {
	client        *s3.Client
	presigner     *s3.PresignClient
	mediaBucket   string
	privateBucket string
	cdnBaseURL    string
	presignMu     sync.RWMutex
	presignCache  map[string]presignedEntry
}

// Object key prefixes that route to the private bucket. Everything else → media.
var privatePrefixes = map[string]struct{}{
	"chat":           {},
	"document-event": {},
}

func NewS3Client() storagesvciface.IStorageService {
	mediaBucket := strings.TrimSpace(os.Getenv("S3_BUCKET_MEDIA"))
	if mediaBucket == "" {
		panic("S3_BUCKET_MEDIA not set")
	}
	privateBucket := strings.TrimSpace(os.Getenv("S3_BUCKET_PRIVATE"))
	if privateBucket == "" {
		panic("S3_BUCKET_PRIVATE not set")
	}
	region := strings.TrimSpace(os.Getenv("S3_REGION"))
	cdnBaseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("S3_CDN_BASE_URL")), "/")
	if cdnBaseURL == "" {
		panic("S3_CDN_BASE_URL not set")
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		panic(fmt.Sprintf("failed to load AWS config: %v", err))
	}

	client := s3.NewFromConfig(cfg)
	return &S3Storage{
		client:        client,
		presigner:     s3.NewPresignClient(client),
		mediaBucket:   mediaBucket,
		privateBucket: privateBucket,
		cdnBaseURL:    cdnBaseURL,
		presignCache:  make(map[string]presignedEntry),
	}
}

func (s *S3Storage) bucketFor(objectName string) string {
	idx := strings.Index(objectName, "/")
	if idx <= 0 {
		return s.mediaBucket
	}
	if _, ok := privatePrefixes[objectName[:idx]]; ok {
		return s.privateBucket
	}
	return s.mediaBucket
}

func (s *S3Storage) Upload(ctx context.Context, objectName string, file multipart.File, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucketFor(objectName)),
		Key:           aws.String(objectName),
		Body:          file,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}
	return nil
}

func (s *S3Storage) GetURL(objectName string) string {
	return fmt.Sprintf("%s/%s", s.cdnBaseURL, objectName)
}

func (s *S3Storage) Delete(ctx context.Context, objectName string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketFor(objectName)),
		Key:    aws.String(objectName),
	})
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}
	return nil
}

func (s *S3Storage) GeneratePresignedURL(ctx context.Context, objectName string, expires time.Duration) (string, error) {
	now := time.Now()
	cacheKey := fmt.Sprintf("%s|%d", objectName, int64(expires))
	s.presignMu.RLock()
	if entry, ok := s.presignCache[cacheKey]; ok && entry.expires.After(now) {
		s.presignMu.RUnlock()
		return entry.url, nil
	}
	s.presignMu.RUnlock()

	req, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketFor(objectName)),
		Key:    aws.String(objectName),
	}, s3.WithPresignExpires(expires))
	if err != nil {
		return "", fmt.Errorf("presign failed: %w", err)
	}

	s.presignMu.Lock()
	for k, v := range s.presignCache {
		if !v.expires.After(now) {
			delete(s.presignCache, k)
		}
	}
	s.presignCache[cacheKey] = presignedEntry{url: req.URL, expires: now.Add(expires - presignSafetyMargin)}
	s.presignMu.Unlock()
	return req.URL, nil
}
