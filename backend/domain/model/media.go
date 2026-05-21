package model

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
)

const (
	megabyte = 1024 * 1024

	MaxAvatarUploadSize    int64 = 5 * megabyte
	MaxBannerUploadSize    int64 = 5 * megabyte
	MaxPostUploadSize      int64 = 5 * megabyte
	MaxChatMediaUploadSize int64 = 20 * megabyte
	MaxDocumentUploadSize  int64 = 20 * megabyte
)

var allowedImageExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".webp": true,
}

var allowedImageContentTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

func ValidateImageExt(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedImageExts[ext] {
		return errors.New("image must be jpg, jpeg, png, or webp")
	}
	return nil
}

func ValidateUploadSize(header *multipart.FileHeader, maxBytes int64) error {
	if header.Size > maxBytes {
		return fmt.Errorf("file must be %d MB or less", maxBytes/megabyte)
	}
	return nil
}

func DetectMultipartContentType(file multipart.File) (string, error) {
	var buf [512]byte
	n, err := file.Read(buf[:])
	if err != nil && err != io.EOF {
		return "", err
	}
	if _, seekErr := file.Seek(0, io.SeekStart); seekErr != nil {
		return "", seekErr
	}
	if n == 0 {
		return "", errors.New("file is empty")
	}
	return strings.ToLower(strings.TrimSpace(strings.Split(http.DetectContentType(buf[:n]), ";")[0])), nil
}

func ValidateImageContentType(contentType string) error {
	if !allowedImageContentTypes[contentType] {
		return fmt.Errorf("image content type %s is not allowed", contentType)
	}
	return nil
}

func ValidateImageFile(file multipart.File) error {
	contentType, err := DetectMultipartContentType(file)
	if err != nil {
		return err
	}
	return ValidateImageContentType(contentType)
}

func MaxImageUploadSize(uploadType string) int64 {
	switch strings.ToLower(uploadType) {
	case "avatar", "community_avatar", "group_avatar":
		return MaxAvatarUploadSize
	case "banner", "community_banner", "event_banner":
		return MaxBannerUploadSize
	default:
		return MaxPostUploadSize
	}
}
