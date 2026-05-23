package helper

import (
	"context"
	"mime/multipart"

	mediamodel "github.com/sunibgram/domain/model/media"
)

type StorageUploader interface {
	ResolveBucketAndPrefix(uploadType string) (string, string, error)
	Upload(ctx context.Context, prefix string, file multipart.File, header *multipart.FileHeader) (string, error)
}

func UploadImage(ctx context.Context, storage StorageUploader, header *multipart.FileHeader, uploadType string) (string, error) {
	if err := mediamodel.ValidateUploadSize(header, mediamodel.MaxImageUploadSize(uploadType)); err != nil {
		return "", err
	}
	if err := mediamodel.ValidateImageExt(header.Filename); err != nil {
		return "", err
	}
	_, prefix, err := storage.ResolveBucketAndPrefix(uploadType)
	if err != nil {
		return "", err
	}
	f, err := header.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()
	if err := mediamodel.ValidateImageFile(f); err != nil {
		return "", err
	}
	objName, err := storage.Upload(ctx, prefix, f, header)
	if err != nil {
		return "", err
	}
	return objName, nil
}
