package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type NotificationService struct {
	notificationRepo repository.INotificationRepository
	userRepo         repository.IUserRepository
	profileRepo      repository.IUserProfileRepository
	userFacultyRepo  repository.IUserFacultyRepository
	storageUsecase   *StorageUsecase
}

func NewNotificationService(
	notificationRepo repository.INotificationRepository,
	userRepo repository.IUserRepository,
	profileRepo repository.IUserProfileRepository,
	userFacultyRepo repository.IUserFacultyRepository,
	storageUsecase *StorageUsecase,
) service.INotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		userRepo:         userRepo,
		profileRepo:      profileRepo,
		userFacultyRepo:  userFacultyRepo,
		storageUsecase:   storageUsecase,
	}
}

func (s *NotificationService) SendNotification(input service.CreateNotificationInput) error {
	input.Type = strings.TrimSpace(input.Type)
	input.Message = strings.TrimSpace(input.Message)
	if input.Type == "" {
		return errors.New("notification type is required")
	}
	if input.Message == "" {
		return errors.New("message is required")
	}
	if input.SenderUserID == 0 || input.RecipientUserID == 0 {
		return errors.New("sender_user_id and recipient_user_id are required")
	}
	if _, err := s.userRepo.FindByID(input.RecipientUserID); err != nil {
		return errors.New("recipient user not found")
	}

	if _, err := s.senderProfile(input.SenderUserID); err != nil {
		return err
	}

	notification := &model.Notification{
		Type:            input.Type,
		Message:         input.Message,
		SenderUserID:    input.SenderUserID,
		ActorUserID:     input.SenderUserID,
		RecipientUserID: input.RecipientUserID,
		ReferenceID:     input.ReferenceID,
		IsRead:          false,
	}
	return s.notificationRepo.Save(notification)
}

func (s *NotificationService) GetNotificationsByRecipientUserID(recipientUserID string, limit, offset int) ([]service.NotificationListItem, error) {
	if recipientUserID == "" {
		return nil, errors.New("recipient user id is required")
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	notifications, err := s.notificationRepo.GetNotificationsByRecipientUserID(context.Background(), recipientUserID, limit, offset)
	if err != nil {
		return nil, err
	}

	items := make([]service.NotificationListItem, len(notifications))
	for i := range notifications {
		actor, err := s.senderProfile(notifications[i].SenderUserID)
		if err != nil {
			return nil, err
		}
		items[i] = service.NotificationListItem{
			Notification: &notifications[i],
			Actor:        actor,
		}
	}
	return items, nil
}

func (s *NotificationService) DeleteNotification(id, requesterID uint, isAdmin bool) error {
	notification, err := s.notificationRepo.FindByID(id)
	if err != nil {
		return errors.New("notification not found")
	}
	if !isAdmin && notification.SenderUserID != requesterID && notification.RecipientUserID != requesterID {
		return errors.New("forbidden")
	}
	return s.notificationRepo.Delete(id)
}

func (s *NotificationService) senderProfile(userID uint) (*model.NotificationSender, error) {
	if sender, err := s.notificationRepo.GetSenderCache(userID); err == nil {
		return sender, nil
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("sender user not found")
	}

	var avatarURL string
	profile, _ := s.profileRepo.FindByUserID(userID)
	if profile != nil && profile.AvatarBucket != "" && profile.AvatarObjectName != "" {
		avatarURL = s.storageUsecase.GetURL(profile.AvatarBucket, profile.AvatarObjectName)
	}

	var faculty *model.Faculty
	userFaculty, _ := s.userFacultyRepo.FindByUserID(userID)
	if userFaculty != nil {
		faculty = &userFaculty.Faculty
	}

	sender := &model.NotificationSender{
		UserID:    user.UserID,
		Username:  user.UserName,
		Name:      user.Name,
		AvatarURL: avatarURL,
		IsPremium: user.IsPremium,
		Faculty:   faculty,
	}
	_ = s.notificationRepo.SetSenderCache(sender)
	return sender, nil
}
