package service

import (
	"context"
	"mime/multipart"

	"github.com/sunibgram/domain/model"
)

type IChatMediaService interface {
	Upload(ctx context.Context, chatID, senderID uint, files []*multipart.FileHeader) ([]model.ChatMedia, error)
	GetByChatID(chatID uint) ([]model.ChatMedia, error)
	GetByChatIDs(chatIDs []uint) ([]model.ChatMedia, error)
	GetURL(bucket, objectName string) string
	GeneratePresignedURL(ctx context.Context, bucket, objectName string) (string, error)
}
