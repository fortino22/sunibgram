package chatroomuc

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strings"
	"unicode/utf8"

	"github.com/sunibgram/application/helper"
	storageuc "github.com/sunibgram/application/usecase/storage"
	chatroommodel "github.com/sunibgram/domain/model/chat_room"
	chatroomrepoiface "github.com/sunibgram/domain/repository/chat_room"
	userrepoiface "github.com/sunibgram/domain/repository/user"
	userblockrepoiface "github.com/sunibgram/domain/repository/user_block"
	chatroomsvciface "github.com/sunibgram/domain/service/chat_room"
)

type ChatRoomService struct {
	roomRepo       chatroomrepoiface.IChatRoomRepository
	userRepo       userrepoiface.IUserRepository
	blockRepo      userblockrepoiface.IUserBlockRepository
	storageUsecase *storageuc.StorageUsecase
}

func NewChatRoomService(roomRepo chatroomrepoiface.IChatRoomRepository, userRepo userrepoiface.IUserRepository, blockRepo userblockrepoiface.IUserBlockRepository, storageUsecase *storageuc.StorageUsecase) chatroomsvciface.IChatRoomService {
	return &ChatRoomService{roomRepo: roomRepo, userRepo: userRepo, blockRepo: blockRepo, storageUsecase: storageUsecase}
}

func (s *ChatRoomService) CreatePrivate(ctx context.Context, initiatorID, targetUserID uint) (*chatroommodel.ChatRoom, error) {
	if initiatorID == targetUserID {
		return nil, errors.New("cannot create private room with yourself")
	}
	if _, err := s.userRepo.FindByID(targetUserID); err != nil {
		return nil, errors.New("target user not found")
	}
	// Cannot open a DM with someone in a mutual block relationship.
	if s.blockRepo != nil {
		blocked, err := s.blockRepo.IsBlocked(initiatorID, targetUserID)
		if err != nil {
			return nil, err
		}
		if blocked {
			return nil, errors.New("blocked")
		}
	}
	if room, err := s.roomRepo.FindPrivateRoom(initiatorID, targetUserID); err == nil {
		if _, err := s.roomKey(room); err != nil {
			return nil, err
		}
		return room, nil
	}

	room := &chatroommodel.ChatRoom{Type: chatroommodel.ChatRoomTypePrivate}
	if err := s.assignRoomKey(room); err != nil {
		return nil, err
	}
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

func (s *ChatRoomService) CreateGroup(ctx context.Context, creatorID uint, name string, description *string, memberIDs []uint) (*chatroommodel.ChatRoom, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("name is required")
	}
	if utf8.RuneCountInString(name) > 100 {
		return nil, errors.New("name must be 100 characters or less")
	}
	if description != nil {
		trimmed := strings.TrimSpace(*description)
		if utf8.RuneCountInString(trimmed) > 500 {
			return nil, errors.New("description must be 500 characters or less")
		}
		*description = trimmed
	}

	room := &chatroommodel.ChatRoom{Name: &name, Description: description, Type: chatroommodel.ChatRoomTypeGroup}
	if err := s.assignRoomKey(room); err != nil {
		return nil, err
	}
	if err := s.roomRepo.Save(room); err != nil {
		return nil, err
	}
	if err := s.roomRepo.AddMember(room.RoomID, creatorID); err != nil {
		return nil, err
	}

	seen := map[uint]struct{}{creatorID: {}}
	dedup := make([]uint, 0, len(memberIDs))
	for _, id := range memberIDs {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		dedup = append(dedup, id)
	}
	if len(dedup) > 0 {
		users, err := s.userRepo.FindByIDs(dedup)
		if err != nil {
			return nil, err
		}
		if len(users) != len(dedup) {
			return nil, errors.New("member user not found")
		}
		if err := s.roomRepo.AddMembersBulk(room.RoomID, dedup); err != nil {
			return nil, err
		}
	}
	return room, nil
}

func (s *ChatRoomService) AddMemberToGroup(ctx context.Context, roomID, requesterID uint, newUserIDs []uint) (*chatroommodel.ChatRoom, []uint, error) {
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
	if room.Type != chatroommodel.ChatRoomTypeGroup {
		return nil, []uint{}, errors.New("room is not a group")
	}

	seen := make(map[uint]struct{}, len(newUserIDs))
	dedup := make([]uint, 0, len(newUserIDs))
	for _, id := range newUserIDs {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		dedup = append(dedup, id)
	}
	alreadyMembers := []uint{}
	toAdd := make([]uint, 0, len(dedup))
	if len(dedup) > 0 {
		users, err := s.userRepo.FindByIDs(dedup)
		if err != nil {
			return nil, []uint{}, err
		}
		if len(users) != len(dedup) {
			return nil, []uint{}, errors.New("target user not found")
		}
		for _, id := range dedup {
			if s.roomRepo.IsMember(roomID, id) {
				alreadyMembers = append(alreadyMembers, id)
				continue
			}
			toAdd = append(toAdd, id)
		}
		if len(toAdd) > 0 {
			if err := s.roomRepo.AddMembersBulk(roomID, toAdd); err != nil {
				return nil, []uint{}, err
			}
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
	if room.Type != chatroommodel.ChatRoomTypeGroup {
		return errors.New("room is not a group")
	}
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return errors.New("name is required")
	}
	if utf8.RuneCountInString(trimmedName) > 100 {
		return errors.New("name must be 100 characters or less")
	}
	if description != nil {
		trimmed := strings.TrimSpace(*description)
		if utf8.RuneCountInString(trimmed) > 500 {
			return errors.New("description must be 500 characters or less")
		}
		*description = trimmed
	}

	var avatarURL *string
	if avatarFile != nil && s.storageUsecase != nil {
		objName, err := helper.UploadImage(ctx, s.storageUsecase, avatarFile, "group_avatar")
		if err != nil {
			return fmt.Errorf("avatar upload failed: %w", err)
		}
		url := s.storageUsecase.GetURL(objName)
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
	if room.Type != chatroommodel.ChatRoomTypeGroup {
		return errors.New("room is not a group")
	}
	return s.roomRepo.RemoveMember(roomID, userID)
}

func (s *ChatRoomService) GetGroupMembers(ctx context.Context, roomID, requesterID uint) ([]chatroomsvciface.RoomParticipant, error) {
	if !s.roomRepo.IsMember(roomID, requesterID) {
		return nil, errors.New("forbidden")
	}
	room, err := s.roomRepo.FindByID(roomID)
	if err != nil {
		return nil, errors.New("chat room not found")
	}
	if room.Type != chatroommodel.ChatRoomTypeGroup {
		return nil, errors.New("room is not a group")
	}
	members, err := s.roomRepo.FindMembers(roomID)
	if err != nil {
		return nil, err
	}
	result := make([]chatroomsvciface.RoomParticipant, 0, len(members))
	for _, m := range members {
		// Hide members in a mutual block relationship with the requester.
		if s.blockRepo != nil && m.UserID != requesterID {
			if blocked, bErr := s.blockRepo.IsBlocked(requesterID, m.UserID); bErr == nil && blocked {
				continue
			}
		}
		avatarURL := ""
		if s.storageUsecase != nil && m.AvatarObjectName != "" {
			avatarURL = s.storageUsecase.GetURL(m.AvatarObjectName)
		}
		result = append(result, chatroomsvciface.RoomParticipant{
			UserID:      m.UserID,
			Username:    m.Username,
			Name:        m.Name,
			FacultyName: m.FacultyName,
			AvatarURL:   avatarURL,
		})
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
	if room.Type != chatroommodel.ChatRoomTypeGroup {
		return errors.New("room is not a group")
	}
	return s.roomRepo.DeleteGroup(roomID)
}

// IsPrivateRoomBlocked reports whether roomID is a private room whose other
// member is mutually blocked with userID. Group rooms return false. Used by the
// socket layer to seal join/typing on an existing A↔B DM after a block.
func (s *ChatRoomService) IsPrivateRoomBlocked(roomID, userID uint) bool {
	if s.blockRepo == nil {
		return false
	}
	memberMap, err := s.roomRepo.FindOtherMemberInfo([]uint{roomID}, userID)
	if err != nil {
		return false
	}
	info, ok := memberMap[roomID]
	if !ok {
		return false
	}
	blocked, err := s.blockRepo.IsBlocked(userID, info.UserID)
	return err == nil && blocked
}

func (s *ChatRoomService) GetUserRooms(userID uint, page, limit int) ([]chatroommodel.ChatRoom, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	return s.roomRepo.FindByUserID(userID, page, limit)
}

func (s *ChatRoomService) GetUserRoomsWithParticipants(userID uint, page, limit int) ([]chatroomsvciface.ChatRoomWithParticipant, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	rooms, total, err := s.roomRepo.FindByUserID(userID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	allIDs := make([]uint, len(rooms))
	privateIDs := make([]uint, 0)
	for i, r := range rooms {
		allIDs[i] = r.RoomID
		if r.Type == chatroommodel.ChatRoomTypePrivate {
			privateIDs = append(privateIDs, r.RoomID)
		}
	}

	memberMap := map[uint]chatroomrepoiface.ChatRoomMemberInfo{}
	if len(privateIDs) > 0 {
		memberMap, err = s.roomRepo.FindOtherMemberInfo(privateIDs, userID)
		if err != nil {
			return nil, 0, err
		}
	}

	// Hide private rooms whose other member is in a mutual block relationship
	// with the viewer (mutual hide). Group rooms are unaffected.
	if s.blockRepo != nil {
		filtered := make([]chatroommodel.ChatRoom, 0, len(rooms))
		newAllIDs := make([]uint, 0, len(rooms))
		for _, r := range rooms {
			if r.Type == chatroommodel.ChatRoomTypePrivate {
				if info, ok := memberMap[r.RoomID]; ok {
					if blocked, bErr := s.blockRepo.IsBlocked(userID, info.UserID); bErr == nil && blocked {
						continue
					}
				}
			}
			filtered = append(filtered, r)
			newAllIDs = append(newAllIDs, r.RoomID)
		}
		rooms = filtered
		allIDs = newAllIDs
	}

	muteMap, _ := s.roomRepo.GetMuteStatuses(userID, allIDs)
	if muteMap == nil {
		muteMap = map[uint]bool{}
	}

	lastMessages, _ := s.roomRepo.FindLastMessages(allIDs)
	if lastMessages == nil {
		lastMessages = map[uint]chatroomrepoiface.LastMessageInfo{}
	}

	unreadCounts, _ := s.roomRepo.FindUnreadCounts(userID, allIDs)
	if unreadCounts == nil {
		unreadCounts = map[uint]int{}
	}

	result := make([]chatroomsvciface.ChatRoomWithParticipant, len(rooms))
	for i, r := range rooms {
		result[i] = chatroomsvciface.ChatRoomWithParticipant{ChatRoom: r, IsMuted: muteMap[r.RoomID], UnreadCount: unreadCounts[r.RoomID]}
		if info, ok := memberMap[r.RoomID]; ok {
			avatarURL := ""
			if s.storageUsecase != nil && info.AvatarObjectName != "" {
				avatarURL = s.storageUsecase.GetURL(info.AvatarObjectName)
			}
			result[i].OtherUser = &chatroomsvciface.RoomParticipant{
				UserID:      info.UserID,
				Username:    info.Username,
				Name:        info.Name,
				FacultyName: info.FacultyName,
				AvatarURL:   avatarURL,
			}
		}
		if lm, ok := lastMessages[r.RoomID]; ok {
			s.decryptLastMessage(&r, &lm)
			result[i].LastMessage = &chatroomsvciface.LastMessage{
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

func (s *ChatRoomService) GetRoomForUser(roomID, viewingUserID uint) (*chatroomsvciface.ChatRoomWithParticipant, error) {
	room, err := s.roomRepo.FindByID(roomID)
	if err != nil {
		return nil, errors.New("chat room not found")
	}
	if !s.roomRepo.IsMember(roomID, viewingUserID) {
		return nil, errors.New("forbidden")
	}
	result := &chatroomsvciface.ChatRoomWithParticipant{ChatRoom: *room}
	if room.Type == chatroommodel.ChatRoomTypePrivate {
		memberMap, err := s.roomRepo.FindOtherMemberInfo([]uint{roomID}, viewingUserID)
		if err == nil {
			if info, ok := memberMap[roomID]; ok {
				avatarURL := ""
				if s.storageUsecase != nil && info.AvatarObjectName != "" {
					avatarURL = s.storageUsecase.GetURL(info.AvatarObjectName)
				}
				result.OtherUser = &chatroomsvciface.RoomParticipant{
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

func (s *ChatRoomService) GetByID(roomID, userID uint) (*chatroommodel.ChatRoom, error) {
	room, err := s.roomRepo.FindByID(roomID)
	if err != nil {
		return nil, errors.New("chat room not found")
	}
	if !s.roomRepo.IsMember(roomID, userID) {
		return nil, errors.New("forbidden")
	}
	return room, nil
}

func (s *ChatRoomService) assignRoomKey(room *chatroommodel.ChatRoom) error {
	masterKey, err := helper.ChatMasterKeyFromEnv("CHAT_MASTER_KEY")
	if err != nil {
		return err
	}
	roomKey, err := helper.ChatGenerateKey()
	if err != nil {
		return err
	}
	encryptedKey, err := helper.ChatEncrypt(roomKey, masterKey)
	if err != nil {
		return err
	}
	room.EncryptedKey = &encryptedKey
	return nil
}

func (s *ChatRoomService) roomKey(room *chatroommodel.ChatRoom) ([]byte, error) {
	if room == nil {
		return nil, errors.New("chat room not found")
	}
	masterKey, err := helper.ChatMasterKeyFromEnv("CHAT_MASTER_KEY")
	if err != nil {
		return nil, err
	}
	if room.EncryptedKey != nil && strings.TrimSpace(*room.EncryptedKey) != "" {
		return helper.ChatDecrypt(*room.EncryptedKey, masterKey)
	}
	roomKey, err := helper.ChatGenerateKey()
	if err != nil {
		return nil, err
	}
	encryptedKey, err := helper.ChatEncrypt(roomKey, masterKey)
	if err != nil {
		return nil, err
	}
	if err := s.roomRepo.UpdateEncryptedKey(room.RoomID, encryptedKey); err != nil {
		return nil, err
	}
	room.EncryptedKey = &encryptedKey
	return roomKey, nil
}

func (s *ChatRoomService) decryptLastMessage(room *chatroommodel.ChatRoom, message *chatroomrepoiface.LastMessageInfo) {
	if room == nil || message == nil || message.Message == nil || strings.TrimSpace(*message.Message) == "" {
		return
	}
	roomKey, err := s.roomKey(room)
	if err != nil {
		return
	}
	plaintext, err := helper.ChatDecrypt(*message.Message, roomKey)
	if err != nil {
		return
	}
	value := string(plaintext)
	message.Message = &value
}
