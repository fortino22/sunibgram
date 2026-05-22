package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/sunibgram/application/helper"
	facultymodel "github.com/sunibgram/domain/model/faculty"
	notificationmodel "github.com/sunibgram/domain/model/notification"
	userprofilemodel "github.com/sunibgram/domain/model/user_profile"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type NotificationService struct {
	notificationRepo repository.INotificationRepository
	userRepo         repository.IUserRepository
	profileRepo      repository.IUserProfileRepository
	userFacultyRepo  repository.IUserFacultyRepository
	storageUsecase   *StorageUsecase
	hub              service.IHub
}

func NewNotificationService(
	notificationRepo repository.INotificationRepository,
	userRepo repository.IUserRepository,
	profileRepo repository.IUserProfileRepository,
	userFacultyRepo repository.IUserFacultyRepository,
	storageUsecase *StorageUsecase,
	hub service.IHub,
) service.INotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		userRepo:         userRepo,
		profileRepo:      profileRepo,
		userFacultyRepo:  userFacultyRepo,
		storageUsecase:   storageUsecase,
		hub:              hub,
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

	actor, err := s.senderProfile(input.SenderUserID)
	if err != nil {
		return err
	}

	notification := &notificationmodel.Notification{
		Type:            input.Type,
		Message:         input.Message,
		SenderUserID:    input.SenderUserID,
		RecipientUserID: input.RecipientUserID,
		ReferenceID:     input.ReferenceID,
	}
	if err := s.notificationRepo.Save(notification); err != nil {
		return err
	}

	if s.hub != nil {
		s.hub.EmitToUser(input.RecipientUserID, "notification:new", service.NotificationListItem{
			Notification: notification,
			Actor:        actor,
			Data:         input.Data,
		})
	}
	return nil
}

func (s *NotificationService) GetNotificationsByRecipientUserID(recipientUserID string, limit, offset int) ([]service.NotificationListItem, error) {
	if recipientUserID == "" {
		return nil, errors.New("recipient user id is required")
	}
	_, limit = helper.NormalizePageLimit(1, limit, 20)
	if offset < 0 {
		offset = 0
	}
	notifications, err := s.notificationRepo.GetNotificationsByRecipientUserID(context.Background(), recipientUserID, limit, offset)
	if err != nil {
		return nil, err
	}

	senders, err := s.senderProfilesByUserID(notifications)
	if err != nil {
		return nil, err
	}

	items := make([]service.NotificationListItem, len(notifications))
	for i := range notifications {
		items[i] = service.NotificationListItem{
			Notification: &notifications[i],
			Actor:        senders[notifications[i].SenderUserID],
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

func (s *NotificationService) senderProfilesByUserID(notifications []notificationmodel.Notification) (map[uint]*notificationmodel.NotificationSender, error) {
	ids := make([]uint, 0, len(notifications))
	seen := make(map[uint]struct{}, len(notifications))
	for i := range notifications {
		userID := notifications[i].SenderUserID
		if _, ok := seen[userID]; ok {
			continue
		}
		seen[userID] = struct{}{}
		ids = append(ids, userID)
	}
	if len(ids) == 0 {
		return map[uint]*notificationmodel.NotificationSender{}, nil
	}

	users, err := s.userRepo.FindByIDs(ids)
	if err != nil {
		return nil, err
	}
	profiles, err := s.profileRepo.FindByUserIDs(ids)
	if err != nil {
		return nil, err
	}
	userFaculties, err := s.userFacultyRepo.FindByUserIDs(ids)
	if err != nil {
		return nil, err
	}

	profilesByUserID := make(map[uint]userprofilemodel.UserProfile, len(profiles))
	for i := range profiles {
		profilesByUserID[profiles[i].UserID] = profiles[i]
	}

	facultiesByUserID := make(map[uint]*facultymodel.Faculty, len(userFaculties))
	for i := range userFaculties {
		facultiesByUserID[userFaculties[i].UserID] = &userFaculties[i].Faculty
	}

	senders := make(map[uint]*notificationmodel.NotificationSender, len(users))
	for i := range users {
		var avatarURL string
		if profile, ok := profilesByUserID[users[i].UserID]; ok && profile.AvatarBucket != "" && profile.AvatarObjectName != "" {
			avatarURL = s.storageUsecase.GetURL(profile.AvatarBucket, profile.AvatarObjectName)
		}

		sender := &notificationmodel.NotificationSender{
			UserID:    users[i].UserID,
			Username:  users[i].UserName,
			Name:      users[i].Name,
			AvatarURL: avatarURL,
			IsPremium: users[i].IsPremium,
			Faculty:   facultiesByUserID[users[i].UserID],
		}
		senders[users[i].UserID] = sender
		_ = s.notificationRepo.SetSenderCache(sender)
	}

	for _, id := range ids {
		if _, ok := senders[id]; !ok {
			return nil, errors.New("sender user not found")
		}
	}
	return senders, nil
}

func (s *NotificationService) senderProfile(userID uint) (*notificationmodel.NotificationSender, error) {
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

	var faculty *facultymodel.Faculty
	userFaculty, _ := s.userFacultyRepo.FindByUserID(userID)
	if userFaculty != nil {
		faculty = &userFaculty.Faculty
	}

	sender := &notificationmodel.NotificationSender{
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
