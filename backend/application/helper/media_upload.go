package helper

import (
	"context"
	"mime/multipart"

	mediamodel "github.com/sunibgram/domain/model/media"
)

type StorageUploader interface {
	ResolveBucket(uploadType string) (string, error)
	Upload(ctx context.Context, bucket string, file multipart.File, header *multipart.FileHeader) (string, error)
}

func UploadImage(ctx context.Context, storage StorageUploader, header *multipart.FileHeader, uploadType string) (string, string, error) {
	if err := mediamodel.ValidateUploadSize(header, mediamodel.MaxImageUploadSize(uploadType)); err != nil {
		return "", "", err
	}
	if err := mediamodel.ValidateImageExt(header.Filename); err != nil {
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
	if err := mediamodel.ValidateImageFile(f); err != nil {
		return "", "", err
	}
	objName, err := storage.Upload(ctx, bucket, f, header)
	if err != nil {
		return "", "", err
	}
	return objName, bucket, nil
}
