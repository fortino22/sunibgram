package usecase

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sunibgram/application/helper"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type EventService struct {
	eventRepo       repository.IEventRepository
	eventReviewRepo repository.IEventReviewRepository
	storageUsecase  *StorageUsecase
}

func NewEventService(eventRepo repository.IEventRepository, eventReviewRepo repository.IEventReviewRepository, storageUsecase *StorageUsecase) service.IEventService {
	return &EventService{eventRepo: eventRepo, eventReviewRepo: eventReviewRepo, storageUsecase: storageUsecase}
}

func (s *EventService) Create(ctx context.Context, title, description, location string, eventType string, startTime, endTime time.Time, banner *multipart.FileHeader, attachments []*multipart.FileHeader) (*model.Event, error) {
	if err := validateEventInput(title, startTime, endTime); err != nil {
		return nil, err
	}
	if err := model.ValidateEventPDFCount(len(attachments)); err != nil {
		return nil, err
	}
	eventType = strings.TrimSpace(eventType)
	if eventType == "" {
		eventType = "unofficial"
	}
	if err := model.ValidateEventType(eventType); err != nil {
		return nil, err
	}
	if eventType == "official" && len(attachments) == 0 {
		return nil, errors.New("proposal attachment required for official events")
	}
	if err := validateEventPDFHeaders(attachments); err != nil {
		return nil, err
	}
	event := &model.Event{
		Title:       strings.TrimSpace(title),
		Description: description,
		Location:    location,
		Type:        eventType,
		StartTime:   startTime.UTC(),
		EndTime:     endTime.UTC(),
	}
	if banner != nil {
		objName, bucket, err := s.uploadImage(ctx, banner, "event_banner")
		if err != nil {
			return nil, fmt.Errorf("banner upload failed: %w", err)
		}
		event.BannerObjectName = objName
		event.BannerBucket = bucket
	}
	if eventType == "official" {
		event.Status = "pending"
	} else {
		event.Status = "approved"
	}
	if len(attachments) > 0 {
		pdfs, err := s.uploadPDFs(ctx, attachments, "document_event")
		if err != nil {
			return nil, err
		}
		event.PDFs = pdfs
		event.AttachmentObjectName = pdfs[0].ObjectName
		event.AttachmentBucket = pdfs[0].Bucket
	}
	if err := s.eventRepo.Save(event); err != nil {
		s.deletePDFs(ctx, event.PDFs)
		return nil, err
	}
	return event, nil
}

func (s *EventService) GetByID(id string) (*model.Event, error) {
	if err := validateEventID(id); err != nil {
		return nil, err
	}
	event, err := s.eventRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	return event, nil
}

func (s *EventService) GetAll(limit int) ([]model.Event, []model.Event, []model.Event, error) {
	_, limit = helper.NormalizePageLimit(1, limit, 10)
	return s.eventRepo.FindGroupedApproved(limit)
}

func (s *EventService) GetPending(limit int) ([]model.Event, error) {
	_, limit = helper.NormalizePageLimit(1, limit, 10)
	return s.eventRepo.FindPending(limit)
}

func (s *EventService) Update(ctx context.Context, id, title, description, location string, startTime, endTime time.Time, banner *multipart.FileHeader, clearBanner bool) (*model.Event, error) {
	if err := validateEventID(id); err != nil {
		return nil, err
	}
	if err := validateEventInput(title, startTime, endTime); err != nil {
		return nil, err
	}
	event, err := s.eventRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("event not found")
	}
	event.Title = strings.TrimSpace(title)
	event.Description = description
	event.Location = location
	if event.Status == "declined" || event.Status == "revision" {
		event.Status = "pending"
	}
	event.StartTime = startTime.UTC()
	event.EndTime = endTime.UTC()
	if clearBanner && event.BannerObjectName != "" {
		_ = s.storageUsecase.Delete(ctx, event.BannerBucket, event.BannerObjectName)
		event.BannerObjectName = ""
		event.BannerBucket = ""
	} else if banner != nil {
		if event.BannerObjectName != "" {
			_ = s.storageUsecase.Delete(ctx, event.BannerBucket, event.BannerObjectName)
		}
		objName, bucket, err := s.uploadImage(ctx, banner, "event_banner")
		if err != nil {
			return nil, fmt.Errorf("banner upload failed: %w", err)
		}
		event.BannerObjectName = objName
		event.BannerBucket = bucket
	}
	if err := s.eventRepo.Update(event); err != nil {
		return nil, err
	}
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
		_ = s.storageUsecase.Delete(ctx, event.BannerBucket, event.BannerObjectName)
	}
	s.deletePDFs(ctx, event.PDFs)
	if event.AttachmentObjectName != "" && len(event.PDFs) == 0 {
		_ = s.storageUsecase.Delete(ctx, event.AttachmentBucket, event.AttachmentObjectName)
	}
	return s.eventRepo.Delete(id)
}

func (s *EventService) ApproveEvent(id, adminID string) error {
	if err := validateEventID(id); err != nil {
		return err
	}
	event, err := s.eventRepo.FindByID(id)
	if err != nil {
		return errors.New("event not found")
	}
	if event.Status == "approved" {
		return errors.New("event already approved")
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
	if event.Status == "declined" {
		return errors.New("event already declined")
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

func (s *EventService) GetEventReviews(eventID string) ([]model.EventReview, error) {
	if err := validateEventID(eventID); err != nil {
		return nil, err
	}
	return s.eventReviewRepo.FindByEventID(eventID)
}

func (s *EventService) createEventReview(eventID, adminID, action, note string) error {
	adminIDUint, err := strconv.ParseUint(adminID, 10, 64)
	if err != nil {
		return errors.New("invalid admin id")
	}
	review := &model.EventReview{
		ID:        uuid.NewString(),
		EventID:   eventID,
		AdminID:   uint(adminIDUint),
		Action:    action,
		Note:      note,
		CreatedAt: time.Now(),
	}
	return s.eventReviewRepo.Create(review)
}

func (s *EventService) GetThreads(ctx context.Context, id string, page, limit int) ([]model.Thread, int64, error) {
	if err := validateEventID(id); err != nil {
		return nil, 0, err
	}
	if _, err := s.eventRepo.FindByID(id); err != nil {
		return nil, 0, errors.New("event not found")
	}
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	return s.eventRepo.FindThreads(id, page, limit)
}

func validateEventID(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return errors.New("invalid event id")
	}
	return nil
}

func (s *EventService) GetURL(bucket, objectName string) string {
	return s.storageUsecase.GetURL(bucket, objectName)
}

func (s *EventService) GeneratePresignedAttachmentURL(ctx context.Context, bucket, objectName string) (string, error) {
	return s.storageUsecase.GeneratePresignedURL(ctx, bucket, objectName)
}

func (s *EventService) uploadImage(ctx context.Context, header *multipart.FileHeader, uploadType string) (string, string, error) {
	return helper.UploadImage(ctx, s.storageUsecase, header, uploadType)
}

func (s *EventService) uploadFile(ctx context.Context, header *multipart.FileHeader, uploadType string) (string, string, error) {
	if err := model.ValidateEventAttachmentExt(header.Filename); err != nil {
		return "", "", err
	}
	if err := model.ValidateEventAttachmentContentType(header.Header.Get("Content-Type")); err != nil {
		return "", "", err
	}
	bucket, err := s.storageUsecase.ResolveBucket(uploadType)
	if err != nil {
		return "", "", err
	}
	f, err := header.Open()
	if err != nil {
		return "", "", err
	}
	defer f.Close()
	objName, err := s.storageUsecase.Upload(ctx, bucket, f, header)
	if err != nil {
		return "", "", err
	}
	return objName, bucket, nil
}

func (s *EventService) uploadPDFs(ctx context.Context, headers []*multipart.FileHeader, uploadType string) (model.EventPDFs, error) {
	if err := validateEventPDFHeaders(headers); err != nil {
		return nil, err
	}

	pdfs := make(model.EventPDFs, 0, len(headers))
	for _, header := range headers {
		objName, bucket, err := s.uploadFile(ctx, header, uploadType)
		if err != nil {
			s.deletePDFs(ctx, pdfs)
			return nil, err
		}
		pdfs = append(pdfs, model.EventPDF{ObjectName: objName, Bucket: bucket})
	}
	return pdfs, nil
}

func validateEventPDFHeaders(headers []*multipart.FileHeader) error {
	if err := model.ValidateEventPDFCount(len(headers)); err != nil {
		return err
	}
	for _, header := range headers {
		if err := model.ValidateEventAttachmentExt(header.Filename); err != nil {
			return err
		}
		if err := model.ValidateEventAttachmentContentType(header.Header.Get("Content-Type")); err != nil {
			return err
		}
	}
	return nil
}

func (s *EventService) deletePDFs(ctx context.Context, pdfs model.EventPDFs) {
	for _, pdf := range pdfs {
		_ = s.storageUsecase.Delete(ctx, pdf.Bucket, pdf.ObjectName)
	}
}

func validateEventInput(title string, startTime, endTime time.Time) error {
	if strings.TrimSpace(title) == "" {
		return errors.New("title is required")
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
