package usecase

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strings"

	"github.com/sunibgram/application/helper"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type ChatRoomService struct {
	roomRepo       repository.IChatRoomRepository
	userRepo       repository.IUserRepository
	storageUsecase *StorageUsecase
}

func NewChatRoomService(roomRepo repository.IChatRoomRepository, userRepo repository.IUserRepository, storageUsecase *StorageUsecase) service.IChatRoomService {
	return &ChatRoomService{roomRepo: roomRepo, userRepo: userRepo, storageUsecase: storageUsecase}
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

func (s *ChatRoomService) CreateGroup(ctx context.Context, creatorID uint, name string, description *string, memberIDs []uint) (*model.ChatRoom, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("name is required")
	}

	room := &model.ChatRoom{Name: &name, Description: description, Type: model.ChatRoomTypeGroup}
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

func (s *ChatRoomService) AddMemberToGroup(ctx context.Context, roomID, requesterID uint, newUserIDs []uint) (*model.ChatRoom, []uint, error) {
	if len(newUserIDs) == 0 {
		return nil, []uint{}, errors.New("user_ids required")
	}
	if !s.roomRepo.IsMember(roomID, requesterID) {
		return nil, []uint{}, errors.New("forbidden")
	}
	room, err := s.roomRepo.FindByID(roomID)
	if err != nil {
		return nil, []uint{}, errors.New("chat room not found")
	}
	if room.Type != model.ChatRoomTypeGroup {
		return nil, []uint{}, errors.New("room is not a group")
	}

	seen := make(map[uint]struct{}, len(newUserIDs))
	alreadyMembers := []uint{}
	for _, id := range newUserIDs {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		if _, err := s.userRepo.FindByID(id); err != nil {
			return nil, []uint{}, errors.New("target user not found")
		}
		if s.roomRepo.IsMember(roomID, id) {
			alreadyMembers = append(alreadyMembers, id)
			continue
		}
		if err := s.roomRepo.AddMember(roomID, id); err != nil {
			return nil, []uint{}, err
		}
	}
	room, err = s.roomRepo.FindByID(roomID)
	if err != nil {
		return nil, []uint{}, err
	}
	return room, alreadyMembers, nil
}

func (s *ChatRoomService) UpdateGroup(ctx context.Context, roomID, requesterID uint, name string, description *string, avatarFile *multipart.FileHeader) error {
	if !s.roomRepo.IsMember(roomID, requesterID) {
		return errors.New("forbidden")
	}
	room, err := s.roomRepo.FindByID(roomID)
	if err != nil {
		return errors.New("chat room not found")
	}
	if room.Type != model.ChatRoomTypeGroup {
		return errors.New("room is not a group")
	}
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return errors.New("name is required")
	}

	var avatarURL *string
	if avatarFile != nil && s.storageUsecase != nil {
		bucket, err := s.storageUsecase.ResolveBucket("group_avatar")
		if err != nil {
			return fmt.Errorf("avatar upload failed: %w", err)
		}
		f, err := avatarFile.Open()
		if err != nil {
			return fmt.Errorf("avatar open failed: %w", err)
		}
		defer f.Close()
		objName, err := s.storageUsecase.Upload(ctx, bucket, f, avatarFile)
		if err != nil {
			return fmt.Errorf("avatar upload failed: %w", err)
		}
		url := s.storageUsecase.GetURL(bucket, objName)
		avatarURL = &url
	}

	return s.roomRepo.UpdateGroup(roomID, trimmedName, description, avatarURL)
}

func (s *ChatRoomService) LeaveGroup(ctx context.Context, roomID, userID uint) error {
	if !s.roomRepo.IsMember(roomID, userID) {
		return errors.New("forbidden")
	}
	room, err := s.roomRepo.FindByID(roomID)
	if err != nil {
		return errors.New("chat room not found")
	}
	if room.Type != model.ChatRoomTypeGroup {
		return errors.New("room is not a group")
	}
	return s.roomRepo.RemoveMember(roomID, userID)
}

func (s *ChatRoomService) GetGroupMembers(ctx context.Context, roomID, requesterID uint) ([]service.RoomParticipant, error) {
	if !s.roomRepo.IsMember(roomID, requesterID) {
		return nil, errors.New("forbidden")
	}
	room, err := s.roomRepo.FindByID(roomID)
	if err != nil {
		return nil, errors.New("chat room not found")
	}
	if room.Type != model.ChatRoomTypeGroup {
		return nil, errors.New("room is not a group")
	}
	members, err := s.roomRepo.FindMembers(roomID)
	if err != nil {
		return nil, err
	}
	result := make([]service.RoomParticipant, len(members))
	for i, m := range members {
		avatarURL := ""
		if s.storageUsecase != nil && m.AvatarObjectName != "" && m.AvatarBucket != "" {
			avatarURL = s.storageUsecase.GetURL(m.AvatarBucket, m.AvatarObjectName)
		}
		result[i] = service.RoomParticipant{
			UserID:      m.UserID,
			Username:    m.Username,
			Name:        m.Name,
			FacultyName: m.FacultyName,
			AvatarURL:   avatarURL,
		}
	}
	return result, nil
}

func (s *ChatRoomService) DeleteGroup(ctx context.Context, roomID, requesterID uint) error {
	if !s.roomRepo.IsMember(roomID, requesterID) {
		return errors.New("forbidden")
	}
	room, err := s.roomRepo.FindByID(roomID)
	if err != nil {
		return errors.New("chat room not found")
	}
	if room.Type != model.ChatRoomTypeGroup {
		return errors.New("room is not a group")
	}
	return s.roomRepo.DeleteGroup(roomID)
}

func (s *ChatRoomService) GetUserRooms(userID uint, page, limit int) ([]model.ChatRoom, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	return s.roomRepo.FindByUserID(userID, page, limit)
}

func (s *ChatRoomService) GetUserRoomsWithParticipants(userID uint, page, limit int) ([]service.ChatRoomWithParticipant, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	rooms, total, err := s.roomRepo.FindByUserID(userID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	allIDs := make([]uint, len(rooms))
	privateIDs := make([]uint, 0)
	for i, r := range rooms {
		allIDs[i] = r.RoomID
		if r.Type == model.ChatRoomTypePrivate {
			privateIDs = append(privateIDs, r.RoomID)
		}
	}

	memberMap := map[uint]repository.ChatRoomMemberInfo{}
	if len(privateIDs) > 0 {
		memberMap, err = s.roomRepo.FindOtherMemberInfo(privateIDs, userID)
		if err != nil {
			return nil, 0, err
		}
	}

	muteMap, _ := s.roomRepo.GetMuteStatuses(userID, allIDs)
	if muteMap == nil {
		muteMap = map[uint]bool{}
	}

	lastMessages, _ := s.roomRepo.FindLastMessages(allIDs)
	if lastMessages == nil {
		lastMessages = map[uint]repository.LastMessageInfo{}
	}

	unreadCounts, _ := s.roomRepo.FindUnreadCounts(userID, allIDs)
	if unreadCounts == nil {
		unreadCounts = map[uint]int{}
	}

	result := make([]service.ChatRoomWithParticipant, len(rooms))
	for i, r := range rooms {
		result[i] = service.ChatRoomWithParticipant{ChatRoom: r, IsMuted: muteMap[r.RoomID], UnreadCount: unreadCounts[r.RoomID]}
		if info, ok := memberMap[r.RoomID]; ok {
			avatarURL := ""
			if s.storageUsecase != nil && info.AvatarObjectName != "" && info.AvatarBucket != "" {
				avatarURL = s.storageUsecase.GetURL(info.AvatarBucket, info.AvatarObjectName)
			}
			result[i].OtherUser = &service.RoomParticipant{
				UserID:      info.UserID,
				Username:    info.Username,
				Name:        info.Name,
				FacultyName: info.FacultyName,
				AvatarURL:   avatarURL,
			}
		}
		if lm, ok := lastMessages[r.RoomID]; ok {
			result[i].LastMessage = &service.LastMessage{
				ChatID:    lm.ChatID,
				SenderID:  lm.SenderID,
				Message:   lm.Message,
				GifURL:    lm.GifURL,
				HasMedia:  lm.HasMedia,
				CreatedAt: lm.CreatedAt,
			}
		}
	}
	return result, total, nil
}

func (s *ChatRoomService) GetRoomForUser(roomID, viewingUserID uint) (*service.ChatRoomWithParticipant, error) {
	room, err := s.roomRepo.FindByID(roomID)
	if err != nil {
		return nil, errors.New("chat room not found")
	}
	if !s.roomRepo.IsMember(roomID, viewingUserID) {
		return nil, errors.New("forbidden")
	}
	result := &service.ChatRoomWithParticipant{ChatRoom: *room}
	if room.Type == model.ChatRoomTypePrivate {
		memberMap, err := s.roomRepo.FindOtherMemberInfo([]uint{roomID}, viewingUserID)
		if err == nil {
			if info, ok := memberMap[roomID]; ok {
				avatarURL := ""
				if s.storageUsecase != nil && info.AvatarObjectName != "" && info.AvatarBucket != "" {
					avatarURL = s.storageUsecase.GetURL(info.AvatarBucket, info.AvatarObjectName)
				}
				result.OtherUser = &service.RoomParticipant{
					UserID:      info.UserID,
					Username:    info.Username,
					Name:        info.Name,
					FacultyName: info.FacultyName,
					AvatarURL:   avatarURL,
				}
			}
		}
	}
	return result, nil
}

func (s *ChatRoomService) GetUserRoomIDs(userID uint) ([]uint, error) {
	return s.roomRepo.GetUserRoomIDs(userID)
}

func (s *ChatRoomService) ToggleMute(ctx context.Context, roomID, userID uint) (bool, error) {
	if !s.roomRepo.IsMember(roomID, userID) {
		return false, errors.New("forbidden")
	}
	return s.roomRepo.ToggleMute(roomID, userID)
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
