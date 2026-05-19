package usecase

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type ReportService struct {
	reportRepo    repository.IReportRepository
	violationRepo repository.IViolationRepository
	userRepo      repository.IUserRepository
	threadRepo    repository.IThreadRepository
}

func NewReportService(
	reportRepo repository.IReportRepository,
	violationRepo repository.IViolationRepository,
	userRepo repository.IUserRepository,
	threadRepo repository.IThreadRepository,
) service.IReportService {
	return &ReportService{
		reportRepo:    reportRepo,
		violationRepo: violationRepo,
		userRepo:      userRepo,
		threadRepo:    threadRepo,
	}
}

func (s *ReportService) Create(reporterID uint, reportedUserID, reportedPostID *uint, violationID uuid.UUID, explanation string) error {
	explanation = strings.TrimSpace(explanation)
	if explanation == "" {
		return errors.New("explanation is required")
	}
	if (reportedUserID == nil && reportedPostID == nil) || (reportedUserID != nil && reportedPostID != nil) {
		return errors.New("exactly one reported_user_id or reported_post_id is required")
	}
	if _, err := s.violationRepo.FindByID(violationID); err != nil {
		return errors.New("violation not found")
	}
	if reportedUserID != nil {
		if *reportedUserID == reporterID {
			return errors.New("you cannot report yourself")
		}
		if _, err := s.userRepo.FindByID(*reportedUserID); err != nil {
			return errors.New("reported user not found")
		}
	}
	if reportedPostID != nil {
		if _, err := s.threadRepo.FindByID(*reportedPostID); err != nil {
			return errors.New("reported post not found")
		}
	}
	return s.reportRepo.Save(&model.Report{
		ReporterID:     reporterID,
		ReportedUserID: reportedUserID,
		ReportedPostID: reportedPostID,
		ViolationID:    violationID,
		Explanation:    explanation,
		Status:         model.ReportStatusPending,
	})
}

func (s *ReportService) GetAll() ([]model.Report, error) {
	return s.reportRepo.FindAll()
}

func (s *ReportService) GetByID(id uuid.UUID) (*model.Report, error) {
	report, err := s.reportRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("report not found")
	}
	return report, nil
}

func (s *ReportService) UpdateStatus(id uuid.UUID, status string) error {
	status = strings.TrimSpace(strings.ToLower(status))
	if !isValidReportStatus(status) {
		return errors.New("invalid report status")
	}
	if _, err := s.reportRepo.FindByID(id); err != nil {
		return errors.New("report not found")
	}
	return s.reportRepo.UpdateStatus(id, status)
}

func (s *ReportService) Delete(id uuid.UUID) error {
	if _, err := s.reportRepo.FindByID(id); err != nil {
		return errors.New("report not found")
	}
	return s.reportRepo.Delete(id)
}

func isValidReportStatus(status string) bool {
	switch status {
	case model.ReportStatusPending, model.ReportStatusReviewed, model.ReportStatusResolved, model.ReportStatusRejected:
		return true
	default:
		return false
	}
}
