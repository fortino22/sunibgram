package reportuc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	reportmodel "github.com/sunibgram/domain/model/report"
	usermodel "github.com/sunibgram/domain/model/user"
	userprofilemodel "github.com/sunibgram/domain/model/user_profile"
	cacherepoiface "github.com/sunibgram/domain/repository/cache"
	reportrepoiface "github.com/sunibgram/domain/repository/report"
	threadrepoiface "github.com/sunibgram/domain/repository/thread"
	userrepoiface "github.com/sunibgram/domain/repository/user"
	violationrepoiface "github.com/sunibgram/domain/repository/violation"
	reportsvciface "github.com/sunibgram/domain/service/report"
	userprofilesvciface "github.com/sunibgram/domain/service/user_profile"
)

const reportCacheTTL = 5 * time.Minute

type ReportService struct {
	reportRepo    reportrepoiface.IReportRepository
	violationRepo violationrepoiface.IViolationRepository
	userRepo      userrepoiface.IUserRepository
	threadRepo    threadrepoiface.IThreadRepository
	profileSvc    userprofilesvciface.IUserProfileService
	cacheRepo     cacherepoiface.ICacheRepository
}

func NewReportService(
	reportRepo reportrepoiface.IReportRepository,
	violationRepo violationrepoiface.IViolationRepository,
	userRepo userrepoiface.IUserRepository,
	threadRepo threadrepoiface.IThreadRepository,
	profileSvc userprofilesvciface.IUserProfileService,
	cacheRepo cacherepoiface.ICacheRepository,
) reportsvciface.IReportService {
	return &ReportService{
		reportRepo:    reportRepo,
		violationRepo: violationRepo,
		userRepo:      userRepo,
		threadRepo:    threadRepo,
		profileSvc:    profileSvc,
		cacheRepo:     cacheRepo,
	}
}

func (s *ReportService) Create(reporterID uint, reportedUserID, reportedPostID *uint, violationID uuid.UUID, explanation string) error {
	explanation = strings.TrimSpace(explanation)
	var reportedOwnerID *uint
	if explanation == "" {
		return errors.New("explanation is required")
	}
	if utf8.RuneCountInString(explanation) > 2000 {
		return errors.New("explanation must be 2000 characters or less")
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
		reportedOwnerID = reportedUserID
	}
	if reportedPostID != nil {
		thread, err := s.threadRepo.FindByID(*reportedPostID)
		if err != nil {
			return errors.New("reported post not found")
		}
		reportedOwnerID = &thread.UserID
	}
	if err := s.reportRepo.Save(&reportmodel.Report{
		ReporterID:     reporterID,
		ReportedUserID: reportedUserID,
		ReportedPostID: reportedPostID,
		ViolationID:    violationID,
		Explanation:    explanation,
		Status:         reportmodel.ReportStatusPending,
	}); err != nil {
		return err
	}
	s.invalidateReportCaches(reporterID, reportedOwnerID)
	return nil
}

func (s *ReportService) GetAll() ([]reportmodel.ReportDetail, error) {
	ctx := context.Background()
	key := "reports:all"
	var cached []reportmodel.ReportDetail
	if s.cacheRepo != nil {
		if err := s.cacheRepo.Get(ctx, key, &cached); err == nil {
			return cached, nil
		}
	}
	reports, err := s.reportRepo.FindAll()
	if err != nil {
		return nil, err
	}
	details, err := s.enrichReports(reports)
	if err != nil {
		return nil, err
	}
	if s.cacheRepo != nil {
		_ = s.cacheRepo.Set(ctx, key, details, reportCacheTTL)
	}
	return details, nil
}

func (s *ReportService) GetByID(id uuid.UUID) (*reportmodel.ReportDetail, error) {
	ctx := context.Background()
	key := reportIDKey(id)
	var cached reportmodel.ReportDetail
	if s.cacheRepo != nil {
		if err := s.cacheRepo.Get(ctx, key, &cached); err == nil {
			return &cached, nil
		}
	}
	report, err := s.reportRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("report not found")
	}
	details, err := s.enrichReports([]reportmodel.Report{*report})
	if err != nil {
		return nil, err
	}
	if len(details) == 0 {
		return nil, errors.New("report not found")
	}
	if s.cacheRepo != nil {
		if err := s.cacheRepo.Set(ctx, key, details[0], reportCacheTTL); err == nil {
			_ = s.cacheRepo.RegisterReportsID(ctx, key)
		}
	}
	return &details[0], nil
}

func (s *ReportService) UpdateStatus(id uuid.UUID, status string) error {
	status = strings.TrimSpace(strings.ToLower(status))
	if !isValidReportStatus(status) {
		return errors.New("invalid report status")
	}
	if _, err := s.reportRepo.FindByID(id); err != nil {
		return errors.New("report not found")
	}
	if err := s.reportRepo.UpdateStatus(id, status); err != nil {
		return err
	}
	report, _ := s.reportRepo.FindByID(id)
	s.invalidateReportCachesFromReport(report)
	return nil
}

func (s *ReportService) Delete(id uuid.UUID) error {
	report, err := s.reportRepo.FindByID(id)
	if err != nil {
		return errors.New("report not found")
	}
	if err := s.reportRepo.Delete(id); err != nil {
		return err
	}
	s.invalidateReportCachesFromReport(report)
	return nil
}

func isValidReportStatus(status string) bool {
	switch status {
	case reportmodel.ReportStatusPending, reportmodel.ReportStatusReviewed, reportmodel.ReportStatusResolved, reportmodel.ReportStatusRejected:
		return true
	default:
		return false
	}
}

func (s *ReportService) enrichReports(reports []reportmodel.Report) ([]reportmodel.ReportDetail, error) {
	details := make([]reportmodel.ReportDetail, len(reports))
	if len(reports) == 0 {
		return details, nil
	}
	userIDs := make([]uint, 0, len(reports)*2)
	seen := make(map[uint]struct{})
	reportedByReportID := make(map[uuid.UUID]uint, len(reports))
	for i := range reports {
		report := reports[i]
		details[i] = reportmodel.ReportDetail{Report: report}
		userIDs = addReportUserID(userIDs, seen, report.ReporterID)
		if report.ReportedUserID != nil {
			reportedByReportID[report.ID] = *report.ReportedUserID
			userIDs = addReportUserID(userIDs, seen, *report.ReportedUserID)
			continue
		}
		if report.ReportedPostID != nil {
			thread, err := s.threadRepo.FindByID(*report.ReportedPostID)
			if err == nil {
				reportedByReportID[report.ID] = thread.UserID
				userIDs = addReportUserID(userIDs, seen, thread.UserID)
			}
		}
	}
	users, err := s.profileSvc.GetUsers(userIDs)
	if err != nil {
		return nil, err
	}
	profiles, _ := s.profileSvc.GetProfiles(userIDs)
	faculties, _ := s.profileSvc.GetUserFaculties(userIDs)
	userInfo := s.buildReportUserInfo(users, profiles, faculties)
	for i := range details {
		report := &details[i].Report
		if u, ok := userInfo[report.ReporterID]; ok {
			details[i].ReporterUser = u
		}
		if reportedID, ok := reportedByReportID[report.ID]; ok {
			if u, ok := userInfo[reportedID]; ok {
				details[i].ReportedUser = u
			}
		}
	}
	return details, nil
}

func addReportUserID(ids []uint, seen map[uint]struct{}, id uint) []uint {
	if id == 0 {
		return ids
	}
	if _, ok := seen[id]; ok {
		return ids
	}
	seen[id] = struct{}{}
	return append(ids, id)
}

func (s *ReportService) buildReportUserInfo(users []usermodel.User, profiles []userprofilemodel.UserProfile, faculties []usermodel.UserFaculty) map[uint]*reportmodel.ReportUserInfo {
	profileByUser := make(map[uint]userprofilemodel.UserProfile, len(profiles))
	for _, profile := range profiles {
		profileByUser[profile.UserID] = profile
	}
	facultyByUser := make(map[uint]usermodel.UserFaculty, len(faculties))
	for _, faculty := range faculties {
		facultyByUser[faculty.UserID] = faculty
	}
	result := make(map[uint]*reportmodel.ReportUserInfo, len(users))
	for _, user := range users {
		info := &reportmodel.ReportUserInfo{
			UserID:    user.UserID,
			Username:  user.UserName,
			IsPremium: user.IsPremium,
		}
		if profile, ok := profileByUser[user.UserID]; ok {
			info.UserAvatar = buildReportAvatarURL(s.profileSvc, profile)
		}
		if uf, ok := facultyByUser[user.UserID]; ok {
			info.Faculty = &reportmodel.ReportUserFaculty{
				FacultyID:   uf.FacultyID,
				FacultyName: uf.Faculty.FacultyName,
			}
		}
		result[user.UserID] = info
	}
	return result
}

func buildReportAvatarURL(profileSvc userprofilesvciface.IUserProfileService, profile userprofilemodel.UserProfile) string {
	if profileSvc == nil || profile.AvatarBucket == "" || profile.AvatarObjectName == "" {
		return ""
	}
	return profileSvc.GetURL(profile.AvatarBucket, profile.AvatarObjectName)
}

func reportIDKey(id uuid.UUID) string {
	return fmt.Sprintf("reports:id:%s", id.String())
}

func (s *ReportService) invalidateReportCachesFromReport(report *reportmodel.Report) {
	if report == nil {
		s.invalidateReportCaches(0, nil)
		return
	}
	reportedUserID := report.ReportedUserID
	if reportedUserID == nil && report.ReportedPostID != nil {
		if thread, err := s.threadRepo.FindByID(*report.ReportedPostID); err == nil {
			reportedUserID = &thread.UserID
		}
	}
	s.invalidateReportCaches(report.ReporterID, reportedUserID)
	if s.cacheRepo != nil {
		_ = s.cacheRepo.Delete(context.Background(), reportIDKey(report.ID))
	}
}

func (s *ReportService) invalidateReportCaches(reporterID uint, reportedUserID *uint) {
	if s.cacheRepo == nil {
		return
	}
	ctx := context.Background()
	_ = s.cacheRepo.Delete(ctx, "reports:all")
	_ = s.cacheRepo.InvalidateReportsID(ctx)
	ids := []uint{reporterID}
	if reportedUserID != nil {
		ids = append(ids, *reportedUserID)
	}
	for _, id := range ids {
		if id == 0 {
			continue
		}
		_ = s.cacheRepo.Delete(ctx,
			fmt.Sprintf("user:%d:avatar", id),
			fmt.Sprintf("profile:user:%d", id),
			fmt.Sprintf("user:id:%d", id),
			fmt.Sprintf("userfaculty:user:%d", id),
		)
	}
}
