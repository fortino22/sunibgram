package helper

import (
	"context"
	"mime/multipart"

	"github.com/sunibgram/domain/model"
)

type StorageUploader interface {
	ResolveBucket(uploadType string) (string, error)
	Upload(ctx context.Context, bucket string, file multipart.File, header *multipart.FileHeader) (string, error)
}

func UploadImage(ctx context.Context, storage StorageUploader, header *multipart.FileHeader, uploadType string) (string, string, error) {
	if err := model.ValidateUploadSize(header, model.MaxImageUploadSize(uploadType)); err != nil {
		return "", "", err
	}
	if err := model.ValidateImageExt(header.Filename); err != nil {
		return "", "", err
	}
	bucket, err := storage.ResolveBucket(uploadType)
	if err != nil {
		return "", "", err
	}
	f, err := header.Open()
	if err != nil {
		return "", "", err
	}
	defer f.Close()
	if err := model.ValidateImageFile(f); err != nil {
		return "", "", err
	}
	objName, err := storage.Upload(ctx, bucket, f, header)
	if err != nil {
		return "", "", err
	}
	return objName, bucket, nil
}
