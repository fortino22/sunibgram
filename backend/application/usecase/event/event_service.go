package eventuc

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/sunibgram/application/helper"
	storageuc "github.com/sunibgram/application/usecase/storage"
	eventreviewmodel "github.com/sunibgram/domain/model/event_review"
	facultylocationmodel "github.com/sunibgram/domain/model/faculty_location"
	mediamodel "github.com/sunibgram/domain/model/media"
	threadmodel "github.com/sunibgram/domain/model/thread"
	eventrepoiface "github.com/sunibgram/domain/repository/event"
	eventparticipantrepoiface "github.com/sunibgram/domain/repository/event_participant"
	eventpinrepoiface "github.com/sunibgram/domain/repository/event_pin"
	eventreviewrepoiface "github.com/sunibgram/domain/repository/event_review"
	facultylocationrepoiface "github.com/sunibgram/domain/repository/faculty_location"
	followrepoiface "github.com/sunibgram/domain/repository/follow"
	threadrepoiface "github.com/sunibgram/domain/repository/thread"
	usersettingsrepoiface "github.com/sunibgram/domain/repository/user_settings"
	eventsvciface "github.com/sunibgram/domain/service/event"
	eventactionconfirmationsvciface "github.com/sunibgram/domain/service/event_action_confirmation"
)

type EventService struct {
	eventRepo           eventrepoiface.IEventRepository
	eventReviewRepo     eventreviewrepoiface.IEventReviewRepository
	pinRepo             eventpinrepoiface.IEventPinRepository
	partRepo            eventparticipantrepoiface.IEventParticipantRepository
	storageUsecase      *storageuc.StorageUsecase
	settingsRepo        usersettingsrepoiface.IUserSettingsRepository
	followRepo          followrepoiface.IFollowRepository
	threadRepo          threadrepoiface.IThreadRepository
	auditSvc            eventactionconfirmationsvciface.IEventActionConfirmationService
	facultyLocationRepo facultylocationrepoiface.IFacultyLocationRepository
}

func NewEventService(
	eventRepo eventrepoiface.IEventRepository,
	eventReviewRepo eventreviewrepoiface.IEventReviewRepository,
	pinRepo eventpinrepoiface.IEventPinRepository,
	partRepo eventparticipantrepoiface.IEventParticipantRepository,
	storageUsecase *storageuc.StorageUsecase,
	settingsRepo usersettingsrepoiface.IUserSettingsRepository,
	followRepo followrepoiface.IFollowRepository,
	threadRepo threadrepoiface.IThreadRepository,
	auditSvc eventactionconfirmationsvciface.IEventActionConfirmationService,
	facultyLocationRepo facultylocationrepoiface.IFacultyLocationRepository,
) eventsvciface.IEventService {
	return &EventService{
		eventRepo:           eventRepo,
		eventReviewRepo:     eventReviewRepo,
		pinRepo:             pinRepo,
		partRepo:            partRepo,
		storageUsecase:      storageUsecase,
		settingsRepo:        settingsRepo,
		followRepo:          followRepo,
		threadRepo:          threadRepo,
		auditSvc:            auditSvc,
		facultyLocationRepo: facultyLocationRepo,
	}
}

// validateAudience requires at least one campus and that every id is a real
// faculty_location (incl. the "All Regions" sentinel). Returns the de-duplicated
// id list.
func (s *EventService) validateAudience(ids []uint) ([]uint, error) {
	seen := map[uint]bool{}
	out := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id == 0 || seen[id] {
			continue
		}
		seen[id] = true
		ok, err := s.facultyLocationRepo.Exists(id)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, errors.New("invalid audience region")
		}
		out = append(out, id)
	}
	if len(out) == 0 {
		return nil, errors.New("at least one audience region is required")
	}
	return out, nil
}

func (s *EventService) GetAudienceByEventIDs(eventIDs []string) (map[string][]facultylocationmodel.FacultyLocation, error) {
	return s.eventRepo.FindAudienceByEventIDs(eventIDs)
}

func (s *EventService) invalidateUserScopedCaches(eventID string) {
	if s.pinRepo != nil {
		s.pinRepo.InvalidateByEvent(eventID)
	}
	if s.partRepo != nil {
		s.partRepo.InvalidateByEvent(eventID)
	}
}

func (s *EventService) Create(ctx context.Context, userID uint, title, description, location string, eventType string, startTime, endTime time.Time, banner *multipart.FileHeader, attachments []*multipart.FileHeader, facultyLocationIDs []uint) (*threadmodel.Event, error) {
	if err := validateEventInput(title, startTime, endTime); err != nil {
		return nil, err
	}
	audience, err := s.validateAudience(facultyLocationIDs)
	if err != nil {
		return nil, err
	}
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	if utf8.RuneCountInString(description) > 2000 {
		return nil, errors.New("description must be 2000 characters or less")
	}
	location = strings.TrimSpace(location)
	if utf8.RuneCountInString(location) > 255 {
		return nil, errors.New("location must be 255 characters or less")
	}
	eventType = strings.TrimSpace(eventType)
	if eventType == "" {
		eventType = "unofficial"
	}
	if err := threadmodel.ValidateEventType(eventType); err != nil {
		return nil, err
	}
	if eventType == "official" {
		if err := threadmodel.ValidateEventDocumentRequired(len(attachments)); err != nil {
			return nil, err
		}
	} else if len(attachments) > 1 {
		return nil, errors.New("only one document is allowed per event")
	}
	if err := validateEventDocHeaders(attachments); err != nil {
		return nil, err
	}
	event := &threadmodel.Event{
		UserID:      userID,
		Title:       title,
		Description: description,
		Location:    location,
		Type:        eventType,
		StartTime:   startTime.UTC(),
		EndTime:     endTime.UTC(),
	}
	if banner != nil {
		objName, err := s.uploadImage(ctx, banner, "event_banner")
		if err != nil {
			return nil, fmt.Errorf("banner upload failed: %w", err)
		}
		event.BannerObjectName = objName
	}
	if eventType == "official" {
		event.Status = "pending"
	} else {
		event.Status = "approved"
	}
	if len(attachments) > 0 {
		docs, err := s.uploadDoc(ctx, attachments[0], "document_event")
		if err != nil {
			return nil, err
		}
		event.Docs = docs
		event.AttachmentObjectName = docs[0].ObjectName
	}
	if err := s.eventRepo.Save(event); err != nil {
		s.deleteDocs(ctx, event.Docs)
		return nil, err
	}
	if err := s.eventRepo.ReplaceAudience(event.ID, audience); err != nil {
		return nil, err
	}
	return event, nil
}

func (s *EventService) GetByID(id string) (*threadmodel.Event, error) {
	if err := validateEventID(id); err != nil {
		return nil, err
	}
	event, err := s.eventRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	return event, nil
}

func (s *EventService) GetByIDWithReviews(id string) (*threadmodel.Event, []eventreviewmodel.EventReview, error) {
	event, err := s.GetByID(id)
	if err != nil {
		return nil, nil, err
	}
	reviews, err := s.eventReviewRepo.FindByEventIDs([]string{event.ID})
	if err != nil {
		return nil, nil, err
	}
	return event, reviews[event.ID], nil
}

func (s *EventService) GetAll(limit int, locationID *uint, region string) ([]threadmodel.Event, []threadmodel.Event, []threadmodel.Event, []threadmodel.Event, error) {
	_, limit = helper.NormalizePageLimit(1, limit, 10)
	return s.eventRepo.FindGroupedApproved(limit, locationID, region)
}

func (s *EventService) GetTopRunningOrFuture(limit int) ([]threadmodel.Event, error) {
	if limit <= 0 {
		limit = 3
	}
	return s.eventRepo.FindTopRunningOrFuture(limit)
}

func (s *EventService) GetPending(limit int) ([]threadmodel.Event, error) {
	_, limit = helper.NormalizePageLimit(1, limit, 10)
	return s.eventRepo.FindPending(limit)
}

func (s *EventService) GetAdminEvents(status, name string, page, limit int) ([]threadmodel.Event, map[string][]eventreviewmodel.EventReview, int64, error) {
	status = strings.TrimSpace(status)
	if err := validateEventStatusFilter(status); err != nil {
		return nil, nil, 0, err
	}
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	events, total, err := s.eventRepo.FindAdmin(status, strings.TrimSpace(name), page, limit)
	if err != nil {
		return nil, nil, 0, err
	}
	reviews, err := s.eventReviewsByEvents(events)
	if err != nil {
		return nil, nil, 0, err
	}
	return events, reviews, total, nil
}

func (s *EventService) GetMyEvents(userID uint, status, name string, page, limit int) ([]threadmodel.Event, map[string][]eventreviewmodel.EventReview, int64, error) {
	status = strings.TrimSpace(status)
	if err := validateEventStatusFilter(status); err != nil {
		return nil, nil, 0, err
	}
	name = strings.TrimSpace(name)
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	events, total, err := s.eventRepo.FindByUserID(userID, status, name, page, limit)
	if err != nil {
		return nil, nil, 0, err
	}
	reviews, err := s.eventReviewsByEvents(events)
	if err != nil {
		return nil, nil, 0, err
	}
	return events, reviews, total, nil
}

func (s *EventService) eventReviewsByEvents(events []threadmodel.Event) (map[string][]eventreviewmodel.EventReview, error) {
	eventIDs := make([]string, 0, len(events))
	for _, event := range events {
		eventIDs = append(eventIDs, event.ID)
	}
	return s.eventReviewRepo.FindByEventIDs(eventIDs)
}

func (s *EventService) Update(ctx context.Context, id, title, description, location, eventType string, startTime, endTime time.Time, banner *multipart.FileHeader, clearBanner bool, attachments []*multipart.FileHeader, facultyLocationIDs []uint) (*threadmodel.Event, error) {
	if err := validateEventID(id); err != nil {
		return nil, err
	}
	if err := validateEventInput(title, startTime, endTime); err != nil {
		return nil, err
	}
	// Audience is optional on update; when provided it fully replaces the set.
	var audience []uint
	if facultyLocationIDs != nil {
		validated, err := s.validateAudience(facultyLocationIDs)
		if err != nil {
			return nil, err
		}
		audience = validated
	}
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	if utf8.RuneCountInString(description) > 2000 {
		return nil, errors.New("description must be 2000 characters or less")
	}
	location = strings.TrimSpace(location)
	if utf8.RuneCountInString(location) > 255 {
		return nil, errors.New("location must be 255 characters or less")
	}
	event, err := s.eventRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("event not found")
	}
	eventType = strings.TrimSpace(eventType)
	if event.Status != "revision" {
		return nil, errors.New("event can only be edited when status is revision")
	}
	if eventType != "" {
		if err := threadmodel.ValidateEventType(eventType); err != nil {
			return nil, err
		}
		event.Type = eventType
	}
	event.Title = title
	event.Description = description
	event.Location = location
	if event.Status == "revision" {
		event.Status = "pending"
	}
	event.StartTime = startTime.UTC()
	event.EndTime = endTime.UTC()
	if clearBanner && event.BannerObjectName != "" {
		_ = s.storageUsecase.Delete(ctx, event.BannerObjectName)
		event.BannerObjectName = ""
	} else if banner != nil {
		if event.BannerObjectName != "" {
			_ = s.storageUsecase.Delete(ctx, event.BannerObjectName)
		}
		objName, err := s.uploadImage(ctx, banner, "event_banner")
		if err != nil {
			return nil, fmt.Errorf("banner upload failed: %w", err)
		}
		event.BannerObjectName = objName
	}
	if len(attachments) > 0 {
		docs, err := s.uploadDoc(ctx, attachments[0], "document_event")
		if err != nil {
			return nil, err
		}
		s.deleteDocs(ctx, event.Docs)
		if event.AttachmentObjectName != "" && len(event.Docs) == 0 {
			_ = s.storageUsecase.Delete(ctx, event.AttachmentObjectName)
		}
		event.Docs = docs
		event.AttachmentObjectName = docs[0].ObjectName
	}
	if err := s.eventRepo.Update(event); err != nil {
		return nil, err
	}
	if audience != nil {
		if err := s.eventRepo.ReplaceAudience(event.ID, audience); err != nil {
			return nil, err
		}
	}
	s.invalidateUserScopedCaches(event.ID)
	return event, nil
}

func (s *EventService) Delete(ctx context.Context, id string) error {
	if err := validateEventID(id); err != nil {
		return err
	}
	event, err := s.eventRepo.FindByID(id)
	if err != nil {
		return errors.New("event not found")
	}
	if event.BannerObjectName != "" {
		_ = s.storageUsecase.Delete(ctx, event.BannerObjectName)
	}
	s.deleteDocs(ctx, event.Docs)
	if event.AttachmentObjectName != "" && len(event.Docs) == 0 {
		_ = s.storageUsecase.Delete(ctx, event.AttachmentObjectName)
	}
	s.invalidateUserScopedCaches(id)
	if err := s.eventRepo.Delete(id); err != nil {
		return err
	}
	if s.auditSvc != nil {
		_ = s.auditSvc.RecordEventDeletion(ctx, id, event.UserID, event.Title)
	}
	return nil
}

func (s *EventService) ApproveEvent(id, adminID string) error {
	if err := validateEventID(id); err != nil {
		return err
	}
	event, err := s.eventRepo.FindByID(id)
	if err != nil {
		return errors.New("event not found")
	}
	if event.Status != "pending" {
		return errors.New("only pending events can be approved")
	}
	if err := s.eventRepo.UpdateStatus(id, "approved"); err != nil {
		return err
	}
	return s.createEventReview(id, adminID, "approved", "")
}

func (s *EventService) DeclineEvent(id, adminID, reason string) error {
	if err := validateEventID(id); err != nil {
		return err
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return errors.New("reason is required")
	}
	event, err := s.eventRepo.FindByID(id)
	if err != nil {
		return errors.New("event not found")
	}
	if event.Status != "pending" {
		return errors.New("only pending events can be declined")
	}
	if err := s.eventRepo.UpdateStatus(id, "declined"); err != nil {
		return err
	}
	return s.createEventReview(id, adminID, "declined", reason)
}

func (s *EventService) RequestRevision(id, adminID, note string) error {
	if err := validateEventID(id); err != nil {
		return err
	}
	note = strings.TrimSpace(note)
	if note == "" {
		return errors.New("note is required")
	}
	if utf8.RuneCountInString(note) > 1000 {
		return errors.New("note must be 1000 characters or less")
	}
	event, err := s.eventRepo.FindByID(id)
	if err != nil {
		return errors.New("event not found")
	}
	if event.Status == "revision" {
		return errors.New("event already requested revision")
	}
	if err := s.eventRepo.UpdateStatus(id, "revision"); err != nil {
		return err
	}
	return s.createEventReview(id, adminID, "revision", note)
}

func (s *EventService) createEventReview(eventID, adminID, action, note string) error {
	adminIDUint, err := strconv.ParseUint(adminID, 10, 64)
	if err != nil {
		return errors.New("invalid admin id")
	}
	review := &eventreviewmodel.EventReview{
		ID:        uuid.NewString(),
		EventID:   eventID,
		AdminID:   uint(adminIDUint),
		Action:    action,
		Note:      note,
		CreatedAt: time.Now(),
	}
	return s.eventReviewRepo.Create(review)
}

func (s *EventService) GetThreads(ctx context.Context, requesterID *uint, id, filter, sort string, page, limit int) ([]threadmodel.Thread, []*threadmodel.Thread, int64, error) {
	if err := validateEventID(id); err != nil {
		return nil, nil, 0, err
	}
	if err := validateEventThreadFilter(filter); err != nil {
		return nil, nil, 0, err
	}
	if err := validateEventThreadSort(sort); err != nil {
		return nil, nil, 0, err
	}
	event, err := s.eventRepo.FindByID(id)
	if err != nil {
		return nil, nil, 0, errors.New("event not found")
	}
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	threads, total, err := s.eventRepo.FindThreads(id, event.UserID, requesterID, filter, sort, page, limit)
	if err != nil {
		return nil, nil, 0, err
	}
	threads, parents, err := s.threadRepo.AttachReplyParents(threads, requesterID)
	if err != nil {
		return nil, nil, 0, err
	}
	return threads, parents, total, nil
}

func validateEventThreadFilter(filter string) error {
	switch filter {
	case "", "official", "unofficial":
		return nil
	default:
		return errors.New("invalid filter: use official or unofficial")
	}
}

func validateEventThreadSort(sort string) error {
	switch sort {
	case "", "latest", "like", "top":
		return nil
	default:
		return errors.New("invalid sort: use latest, like, or top")
	}
}

func validateEventID(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return errors.New("invalid event id")
	}
	return nil
}

func validateEventStatusFilter(status string) error {
	if status == "" {
		return nil
	}
	switch status {
	case "pending", "approved", "declined", "revision":
		return nil
	default:
		return errors.New("invalid event status")
	}
}

func (s *EventService) GetURL(objectName string) string {
	return s.storageUsecase.GetURL(objectName)
}

func (s *EventService) GeneratePresignedAttachmentURL(ctx context.Context, objectName string) (string, error) {
	return s.storageUsecase.GeneratePresignedURL(ctx, objectName)
}

func (s *EventService) uploadImage(ctx context.Context, header *multipart.FileHeader, uploadType string) (string, error) {
	return helper.UploadImage(ctx, s.storageUsecase, header, uploadType)
}

func (s *EventService) uploadFile(ctx context.Context, header *multipart.FileHeader, uploadType string) (string, error) {
	if err := mediamodel.ValidateUploadSize(header, mediamodel.MaxDocumentUploadSize); err != nil {
		return "", err
	}
	if err := threadmodel.ValidateEventDocExt(header.Filename); err != nil {
		return "", err
	}
	_, prefix, err := s.storageUsecase.ResolveBucketAndPrefix(uploadType)
	if err != nil {
		return "", err
	}
	f, err := header.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()
	if err := threadmodel.ValidateEventDocFile(f); err != nil {
		return "", err
	}
	objName, err := s.storageUsecase.Upload(ctx, prefix, f, header)
	if err != nil {
		return "", err
	}
	return objName, nil
}

func (s *EventService) uploadDoc(ctx context.Context, header *multipart.FileHeader, uploadType string) (threadmodel.EventDocs, error) {
	objName, err := s.uploadFile(ctx, header, uploadType)
	if err != nil {
		return nil, err
	}
	return threadmodel.EventDocs{{ObjectName: objName}}, nil
}

func validateEventDocHeaders(headers []*multipart.FileHeader) error {
	for _, header := range headers {
		if err := mediamodel.ValidateUploadSize(header, mediamodel.MaxDocumentUploadSize); err != nil {
			return err
		}
		if err := threadmodel.ValidateEventDocExt(header.Filename); err != nil {
			return err
		}
	}
	return nil
}

func (s *EventService) deleteDocs(ctx context.Context, docs threadmodel.EventDocs) {
	for _, doc := range docs {
		if doc.ObjectName == "" {
			continue
		}
		_ = s.storageUsecase.Delete(ctx, doc.ObjectName)
	}
}

func validateEventInput(title string, startTime, endTime time.Time) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return errors.New("title is required")
	}
	if utf8.RuneCountInString(title) > 255 {
		return errors.New("title must be 255 characters or less")
	}
	if startTime.IsZero() {
		return errors.New("start_time is required")
	}
	if endTime.IsZero() {
		return errors.New("end_time is required")
	}
	if !endTime.After(startTime) {
		return errors.New("end_time must be after start_time")
	}
	return nil
}
