package reportctrl

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	reportmodel "github.com/sunibgram/domain/model/report"
	threadmodel "github.com/sunibgram/domain/model/thread"
	bookmarksvciface "github.com/sunibgram/domain/service/bookmark"
	likesvciface "github.com/sunibgram/domain/service/like"
	reportsvciface "github.com/sunibgram/domain/service/report"
	threadsvciface "github.com/sunibgram/domain/service/thread"
	threadmediasvciface "github.com/sunibgram/domain/service/thread_media"
	threadpollsvciface "github.com/sunibgram/domain/service/thread_poll"
	userprofilesvciface "github.com/sunibgram/domain/service/user_profile"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	categoryctrl "github.com/sunibgram/webapi/controller/category"
	threadctrl "github.com/sunibgram/webapi/controller/thread"
	userprofilectrl "github.com/sunibgram/webapi/controller/user_profile"
	violationctrl "github.com/sunibgram/webapi/controller/violation"
	utils "github.com/sunibgram/webapi/dto/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
	reportdto "github.com/sunibgram/webapi/dto/report"
	threaddto "github.com/sunibgram/webapi/dto/thread"
)

type ReportController struct {
	reportService   reportsvciface.IReportService
	mediaService    threadmediasvciface.IThreadMediaService
	likeService     likesvciface.ILikeService
	bookmarkService bookmarksvciface.IBookmarkService
	threadService   threadsvciface.IThreadService
	pollService     threadpollsvciface.IThreadPollService
	profileService  userprofilesvciface.IUserProfileService
}

func NewReportController(
	reportService reportsvciface.IReportService,
	mediaService threadmediasvciface.IThreadMediaService,
	likeService likesvciface.ILikeService,
	bookmarkService bookmarksvciface.IBookmarkService,
	threadService threadsvciface.IThreadService,
	pollService threadpollsvciface.IThreadPollService,
	profileService userprofilesvciface.IUserProfileService,
) *ReportController {
	return &ReportController{
		reportService:   reportService,
		mediaService:    mediaService,
		likeService:     likeService,
		bookmarkService: bookmarkService,
		threadService:   threadService,
		pollService:     pollService,
		profileService:  profileService,
	}
}

func (c *ReportController) Create(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	var req reportdto.CreateReportRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid request body"})
	}
	violationID, err := uuid.Parse(req.ViolationID)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid violation id"})
	}
	if err := c.reportService.Create(claims.UserID, req.ReportedUserID, req.ReportedPostID, violationID, req.Explanation); err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "report submitted successfully"})
}

func (c *ReportController) GetAll(ctx *fiber.Ctx) error {
	reports, err := c.reportService.GetAll()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": toReportResponses(reports)})
}

// GetReported returns reported threads (full thread data) and reported users
// (full profile), grouped by entity. Admin-only. Optional ?status= filter
// (pending|reviewed|resolved|rejected); "all" or omitted = every status.
func (c *ReportController) GetReported(ctx *fiber.Ctx) error {
	status := strings.TrimSpace(strings.ToLower(ctx.Query("status", "")))
	if status == "all" {
		status = ""
	}
	if status != "" && !isValidReportStatusFilter(status) {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid report status"})
	}

	result, err := c.reportService.GetReportedEntities(status)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}

	return ctx.Status(fiber.StatusOK).JSON(reportdto.ReportedEntitiesResponse{
		Threads: c.toReportedThreadGroups(result.Threads),
		Users:   c.toReportedUserGroups(result.Users),
	})
}

func isValidReportStatusFilter(status string) bool {
	switch status {
	case reportmodel.ReportStatusPending, reportmodel.ReportStatusReviewed,
		reportmodel.ReportStatusResolved, reportmodel.ReportStatusRejected:
		return true
	}
	return false
}

func (c *ReportController) toReportedThreadGroups(groups []reportmodel.ReportedThreadGroup) []reportdto.ReportedThreadGroupResponse {
	out := make([]reportdto.ReportedThreadGroupResponse, 0, len(groups))
	if len(groups) == 0 {
		return out
	}

	threads := make([]threadmodel.Thread, len(groups))
	for i, g := range groups {
		threads[i] = g.Thread
	}
	feedByID := c.buildFeedItems(threads)

	for _, g := range groups {
		item, ok := feedByID[g.ThreadID]
		if !ok {
			continue
		}
		out = append(out, reportdto.ReportedThreadGroupResponse{
			Thread:      item,
			ReportCount: g.ReportCount,
			Reports:     toReportResponses(g.Reports),
		})
	}
	return out
}

func (c *ReportController) toReportedUserGroups(groups []reportmodel.ReportedUserGroup) []reportdto.ReportedUserGroupResponse {
	out := make([]reportdto.ReportedUserGroupResponse, 0, len(groups))
	for _, g := range groups {
		fp, err := c.profileService.GetProfile(g.UserID, nil)
		if err != nil {
			continue // reported user no longer exists
		}
		out = append(out, reportdto.ReportedUserGroupResponse{
			User:        userprofilectrl.ToUserProfileResponse(fp, c.profileService.GetURL, true),
			ReportCount: g.ReportCount,
			Reports:     toReportResponses(g.Reports),
		})
	}
	return out
}

// buildFeedItems hydrates full FeedItemResponse (media, poll, author) for the
// given threads via the shared FeedBuilder, keyed by thread ID. Viewer is nil
// (admin moderation view) so per-viewer flags resolve to false.
func (c *ReportController) buildFeedItems(threads []threadmodel.Thread) map[uint]threaddto.FeedItemResponse {
	allIDs := make([]uint, 0, len(threads)*2)
	for _, t := range threads {
		allIDs = append(allIDs, t.ThreadID)
		if t.OriginalThread != nil {
			allIDs = append(allIDs, t.OriginalThread.ThreadID)
		}
	}
	mediaByThread := ctrlshared.LoadThreadMediaByID(c.mediaService, allIDs)

	fb := ctrlshared.NewFeedBuilder(c.mediaService, c.likeService, c.bookmarkService, c.threadService, nil, c.profileService, c.pollService, nil)
	fb.LoadAvatarsForUserIDs(threadctrl.ThreadUserIDs(threads))

	pollIDs := make([]uint, 0, len(threads))
	for i := range threads {
		t := &threads[i]
		if t.Type == threadmodel.ThreadTypeRepost && t.OriginalThread != nil {
			pollIDs = append(pollIDs, t.OriginalThread.ThreadID)
		} else {
			pollIDs = append(pollIDs, t.ThreadID)
		}
	}
	fb.LoadPollsForThreadIDs(pollIDs)

	data := make([]threaddto.FeedItemResponse, len(threads))
	for i := range threads {
		t := threads[i]
		mediaID := t.ThreadID
		if t.Type == threadmodel.ThreadTypeRepost && t.OriginalThread != nil {
			mediaID = t.OriginalThread.ThreadID
		}
		data[i] = fb.ToFeedItem(&t, mediaByThread[mediaID])
	}
	ctrlshared.EnsureFalseFeedViewerFlags(data)

	out := make(map[uint]threaddto.FeedItemResponse, len(data))
	for _, item := range data {
		out[item.ID] = item
	}
	return out
}

func (c *ReportController) GetByID(ctx *fiber.Ctx) error {
	id, ok := categoryctrl.ParseUUIDParam(ctx, "id", "report")
	if !ok {
		return nil
	}
	report, err := c.reportService.GetByID(id)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(toReportResponse(report))
}

func (c *ReportController) UpdateStatus(ctx *fiber.Ctx) error {
	id, ok := categoryctrl.ParseUUIDParam(ctx, "id", "report")
	if !ok {
		return nil
	}

	var req reportdto.UpdateReportStatusRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid request body"})
	}
	if err := c.reportService.UpdateStatus(id, req.Status); err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "report status updated successfully"})
}

func (c *ReportController) Delete(ctx *fiber.Ctx) error {
	id, ok := categoryctrl.ParseUUIDParam(ctx, "id", "report")
	if !ok {
		return nil
	}
	if err := c.reportService.Delete(id); err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "report deleted successfully"})
}

func toReportResponse(detail *reportmodel.ReportDetail) reportdto.ReportResponse {
	report := &detail.Report
	response := reportdto.ReportResponse{
		ID:             report.ID.String(),
		ReporterID:     report.ReporterID,
		ReportedUserID: report.ReportedUserID,
		ReportedPostID: report.ReportedPostID,
		ViolationID:    report.ViolationID.String(),
		Explanation:    report.Explanation,
		Status:         report.Status,
		CreatedAt:      utils.Wrap(report.CreatedAt),
		UpdatedAt:      utils.Wrap(report.UpdatedAt),
	}
	if report.Violation.ID != uuid.Nil {
		violation := violationctrl.ToViolationResponse(&report.Violation)
		response.Violation = &violation
	}
	response.ReporterUser = toReportUserResponse(detail.ReporterUser)
	response.ReportedUser = toReportUserResponse(detail.ReportedUser)
	return response
}

func toReportResponses(reports []reportmodel.ReportDetail) []reportdto.ReportResponse {
	responses := make([]reportdto.ReportResponse, len(reports))
	for i := range reports {
		responses[i] = toReportResponse(&reports[i])
	}
	return responses
}

func toReportUserResponse(user *reportmodel.ReportUserInfo) *reportdto.ReportUserResponse {
	if user == nil {
		return nil
	}
	response := &reportdto.ReportUserResponse{
		UserID:     user.UserID,
		Username:   user.Username,
		UserAvatar: user.UserAvatar,
		IsPremium:  user.IsPremium,
	}
	if user.Faculty != nil {
		response.Faculty = &reportdto.ReportFacultyResponse{
			FacultyID:   user.Faculty.FacultyID,
			FacultyName: user.Faculty.FacultyName,
		}
	}
	return response
}
