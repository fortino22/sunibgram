package usecase

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sunibgram/application/helper"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type EventService struct {
	eventRepo      repository.IEventRepository
	storageUsecase *StorageUsecase
}

func NewEventService(eventRepo repository.IEventRepository, storageUsecase *StorageUsecase) service.IEventService {
	return &EventService{eventRepo: eventRepo, storageUsecase: storageUsecase}
}

func (s *EventService) Create(ctx context.Context, title, description, location string, startTime, endTime time.Time, banner *multipart.FileHeader) (*model.Event, error) {
	if err := validateEventInput(title, startTime, endTime); err != nil {
		return nil, err
	}
	event := &model.Event{
		Title:       strings.TrimSpace(title),
		Description: description,
		Location:    location,
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
	if err := s.eventRepo.Save(event); err != nil {
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
	return s.eventRepo.FindGrouped(limit)
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
	return s.eventRepo.Delete(id)
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

func (s *EventService) uploadImage(ctx context.Context, header *multipart.FileHeader, uploadType string) (string, string, error) {
	return helper.UploadImage(ctx, s.storageUsecase, header, uploadType)
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
