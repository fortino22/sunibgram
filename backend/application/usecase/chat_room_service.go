package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type ChatRoomService struct {
	roomRepo repository.IChatRoomRepository
	userRepo repository.IUserRepository
}

func NewChatRoomService(roomRepo repository.IChatRoomRepository, userRepo repository.IUserRepository) service.IChatRoomService {
	return &ChatRoomService{roomRepo: roomRepo, userRepo: userRepo}
}

func (s *ChatRoomService) CreatePrivate(ctx context.Context, initiatorID, targetUserID uint) (*model.ChatRoom, error) {
	if initiatorID == targetUserID {
		return nil, errors.New("cannot create private room with yourself")
	}
	if _, err := s.userRepo.FindByID(targetUserID); err != nil {
		return nil, errors.New("target user not found")
	}
	if room, err := s.roomRepo.FindPrivateRoom(initiatorID, targetUserID); err == nil {
		return room, nil
	}

	room := &model.ChatRoom{Type: model.ChatRoomTypePrivate}
	if err := s.roomRepo.Save(room); err != nil {
		return nil, err
	}
	if err := s.roomRepo.AddMember(room.RoomID, initiatorID); err != nil {
		return nil, err
	}
	if err := s.roomRepo.AddMember(room.RoomID, targetUserID); err != nil {
		return nil, err
	}
	return room, nil
}

func (s *ChatRoomService) CreateGroup(ctx context.Context, creatorID uint, name string, memberIDs []uint) (*model.ChatRoom, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("name is required")
	}

	room := &model.ChatRoom{Name: &name, Type: model.ChatRoomTypeGroup}
	if err := s.roomRepo.Save(room); err != nil {
		return nil, err
	}
	if err := s.roomRepo.AddMember(room.RoomID, creatorID); err != nil {
		return nil, err
	}

	seen := map[uint]struct{}{creatorID: {}}
	for _, id := range memberIDs {
		if _, ok := seen[id]; ok {
			continue
		}
		if _, err := s.userRepo.FindByID(id); err != nil {
			return nil, errors.New("member user not found")
		}
		if err := s.roomRepo.AddMember(room.RoomID, id); err != nil {
			return nil, err
		}
		seen[id] = struct{}{}
	}
	return room, nil
}

func (s *ChatRoomService) GetUserRooms(userID uint, page, limit int) ([]model.ChatRoom, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.roomRepo.FindByUserID(userID, page, limit)
}

func (s *ChatRoomService) GetByID(roomID, userID uint) (*model.ChatRoom, error) {
	room, err := s.roomRepo.FindByID(roomID)
	if err != nil {
		return nil, errors.New("chat room not found")
	}
	if !s.roomRepo.IsMember(roomID, userID) {
		return nil, errors.New("forbidden")
	}
	return room, nil
}
