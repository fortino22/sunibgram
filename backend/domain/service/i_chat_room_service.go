package service

import (
	"context"

	"github.com/sunibgram/domain/model"
)

type IChatRoomService interface {
	CreatePrivate(ctx context.Context, initiatorID, targetUserID uint) (*model.ChatRoom, error)
	CreateGroup(ctx context.Context, creatorID uint, name string, memberIDs []uint) (*model.ChatRoom, error)
	GetUserRooms(userID uint, page, limit int) ([]model.ChatRoom, int64, error)
	GetByID(roomID, userID uint) (*model.ChatRoom, error)
}
