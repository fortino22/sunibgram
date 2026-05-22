package service

import (
	"context"
	"mime/multipart"

	chatmediamodel "github.com/sunibgram/domain/model/chat_media"
)

type IChatMediaService interface {
	Upload(ctx context.Context, chatID, senderID uint, files []*multipart.FileHeader) ([]chatmediamodel.ChatMedia, error)
	GetByChatID(chatID uint) ([]chatmediamodel.ChatMedia, error)
	GetByChatIDs(chatIDs []uint) ([]chatmediamodel.ChatMedia, error)
	GetURL(bucket, objectName string) string
	GeneratePresignedURL(ctx context.Context, bucket, objectName string) (string, error)
}
