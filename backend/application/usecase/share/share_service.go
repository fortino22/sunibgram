package shareuc

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	storageuc "github.com/sunibgram/application/usecase/storage"
	exploremodel "github.com/sunibgram/domain/model/explore"
	chatrepoiface "github.com/sunibgram/domain/repository/chat"
	threadrepoiface "github.com/sunibgram/domain/repository/thread"
	userrepoiface "github.com/sunibgram/domain/repository/user"
	userblockrepoiface "github.com/sunibgram/domain/repository/user_block"
	chatsvciface "github.com/sunibgram/domain/service/chat"
	chatroomsvciface "github.com/sunibgram/domain/service/chat_room"
	sharesvciface "github.com/sunibgram/domain/service/share"
)

type ShareService struct {
	threadRepo     threadrepoiface.IThreadRepository
	userRepo       userrepoiface.IUserRepository
	chatRepo       chatrepoiface.IChatRepository
	blockRepo      userblockrepoiface.IUserBlockRepository
	chatService    chatsvciface.IChatService
	roomService    chatroomsvciface.IChatRoomService
	storageUsecase *storageuc.StorageUsecase
}

func NewShareService(
	threadRepo threadrepoiface.IThreadRepository,
	userRepo userrepoiface.IUserRepository,
	chatRepo chatrepoiface.IChatRepository,
	blockRepo userblockrepoiface.IUserBlockRepository,
	chatService chatsvciface.IChatService,
	roomService chatroomsvciface.IChatRoomService,
	storageUsecase *storageuc.StorageUsecase,
) sharesvciface.IShareService {
	return &ShareService{
		threadRepo:     threadRepo,
		userRepo:       userRepo,
		chatRepo:       chatRepo,
		blockRepo:      blockRepo,
		chatService:    chatService,
		roomService:    roomService,
		storageUsecase: storageUsecase,
	}
}

func (s *ShareService) RecordCopy(ctx context.Context, userID, threadID uint) (int, error) {
	thread, err := s.threadRepo.FindByID(threadID)
	if err != nil {
		return 0, errors.New("thread not found")
	}
	if err := s.checkBlockedAuthor(userID, thread.UserID); err != nil {
		return 0, err
	}
	if err := s.threadRepo.IncrementShareCount(threadID); err != nil {
		return 0, err
	}
	return thread.ShareCount + 1, nil
}

func (s *ShareService) SendToUsers(ctx context.Context, senderID, threadID uint, recipientIDs []uint, message *string) (int, error) {
	if len(recipientIDs) == 0 {
		return 0, errors.New("user_ids required")
	}
	if len(recipientIDs) > 20 {
		return 0, errors.New("cannot share with more than 20 users at once")
	}
	thread, err := s.threadRepo.FindByID(threadID)
	if err != nil {
		return 0, errors.New("thread not found")
	}
	if err := s.checkBlockedAuthor(senderID, thread.UserID); err != nil {
		return 0, err
	}

	var trimmedMessage *string
	if message != nil {
		trimmed := strings.TrimSpace(*message)
		if trimmed != "" {
			if utf8.RuneCountInString(trimmed) > 5000 {
				return 0, errors.New("message must be 5000 characters or less")
			}
			trimmedMessage = &trimmed
		}
	}

	seen := make(map[uint]struct{}, len(recipientIDs))
	dedup := make([]uint, 0, len(recipientIDs))
	for _, id := range recipientIDs {
		if id == senderID {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		dedup = append(dedup, id)
	}
	if len(dedup) == 0 {
		return 0, errors.New("no valid recipients")
	}

	sharedID := thread.ThreadID
	sent := 0
	for _, recipientID := range dedup {
		// Skip recipients in a mutual block relationship with the sender.
		if s.blockRepo != nil {
			if blocked, err := s.blockRepo.IsBlocked(senderID, recipientID); err == nil && blocked {
				continue
			}
		}
		room, err := s.roomService.CreatePrivate(ctx, senderID, recipientID)
		if err != nil {
			continue
		}
		if _, err := s.chatService.Send(ctx, room.RoomID, senderID, trimmedMessage, nil, nil, &sharedID); err != nil {
			continue
		}
		sent++
	}
	if sent == 0 {
		return 0, errors.New("failed to deliver share")
	}
	_ = s.threadRepo.IncrementShareCount(threadID)
	return sent, nil
}

// checkBlockedAuthor rejects sharing a thread when the sharer and the thread's
// author are in a mutual block relationship.
func (s *ShareService) checkBlockedAuthor(senderID, authorID uint) error {
	if s.blockRepo == nil || senderID == 0 || senderID == authorID {
		return nil
	}
	blocked, err := s.blockRepo.IsBlocked(senderID, authorID)
	if err != nil {
		return err
	}
	if blocked {
		return errors.New("blocked")
	}
	return nil
}

func (s *ShareService) SearchUsers(ctx context.Context, query string, requesterID uint) ([]exploremodel.ExploreUserResult, error) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return []exploremodel.ExploreUserResult{}, nil
	}
	return s.userRepo.SearchForShare(ctx, trimmed, requesterID)
}
