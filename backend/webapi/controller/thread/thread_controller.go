package threadctrl

import (
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	communitymodel "github.com/sunibgram/domain/model/community"
	threadmodel "github.com/sunibgram/domain/model/thread"
	threadmediamodel "github.com/sunibgram/domain/model/thread_media"
	bookmarksvciface "github.com/sunibgram/domain/service/bookmark"
	communitymembersvciface "github.com/sunibgram/domain/service/community_member"
	likesvciface "github.com/sunibgram/domain/service/like"
	threadsvciface "github.com/sunibgram/domain/service/thread"
	threadmediasvciface "github.com/sunibgram/domain/service/thread_media"
	threadpinsvciface "github.com/sunibgram/domain/service/thread_pin"
	threadpollsvciface "github.com/sunibgram/domain/service/thread_poll"
	userprofilesvciface "github.com/sunibgram/domain/service/user_profile"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	threadpollctrl "github.com/sunibgram/webapi/controller/thread_poll"
	authdto "github.com/sunibgram/webapi/dto/auth"
	threaddto "github.com/sunibgram/webapi/dto/thread"
	threadmediadto "github.com/sunibgram/webapi/dto/thread_media"
)

type ThreadController struct {
	threadService    threadsvciface.IThreadService
	mediaService     threadmediasvciface.IThreadMediaService
	memberService    communitymembersvciface.ICommunityMemberService
	likeService      likesvciface.ILikeService
	bookmarkService  bookmarksvciface.IBookmarkService
	threadPinService threadpinsvciface.IThreadPinService
	profileService   userprofilesvciface.IUserProfileService
	pollService      threadpollsvciface.IThreadPollService
}

func NewThreadController(threadService threadsvciface.IThreadService, mediaService threadmediasvciface.IThreadMediaService, memberService communitymembersvciface.ICommunityMemberService, likeService likesvciface.ILikeService, bookmarkService bookmarksvciface.IBookmarkService, threadPinService threadpinsvciface.IThreadPinService, profileService userprofilesvciface.IUserProfileService, pollService threadpollsvciface.IThreadPollService) *ThreadController {
	return &ThreadController{threadService: threadService, mediaService: mediaService, memberService: memberService, likeService: likeService, bookmarkService: bookmarkService, threadPinService: threadPinService, profileService: profileService, pollService: pollService}
}

func (c *ThreadController) Create(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	content := ctx.FormValue("content")
	files := ThreadMediaFiles(ctx)
	if err := threadmediamodel.ValidateThreadMediaCount(len(files)); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	var parentID *uint
	if raw := ctx.FormValue("parent_id"); raw != "" {
		if pid, err := strconv.ParseUint(raw, 10, 64); err == nil {
			v := uint(pid)
			parentID = &v
		}
	}

	eventID, _, err := ParseOptionalEventID(ctx)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}

	var replyPermission *string
	if raw := strings.TrimSpace(ctx.FormValue("reply_permission")); raw != "" {
		replyPermission = &raw
	}

	pollInput, _, hasPoll, err := threadpollctrl.ParseCreatePollFromMultipart(ctx)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	if hasPoll && parentID != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "polls are only allowed on post-type threads"})
	}

	thread, err := c.threadService.Create(ctx.Context(), claims.UserID, content, parentID, eventID, replyPermission)
	if err != nil {
		if err.Error() == "not allowed to reply to this thread" || err.Error() == "blocked" {
			return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: err.Error()})
		}
		return ctrlshared.HandleServiceError(ctx, err)
	}

	if len(files) > 0 {
		if _, err := c.mediaService.Upload(ctx.Context(), thread.ThreadID, claims.UserID, files); err != nil {
			_ = c.threadService.Delete(ctx.Context(), thread.ThreadID, claims.UserID)
			return ctrlshared.HandleServiceError(ctx, err)
		}
	}

	if hasPoll {
		if _, err := c.pollService.Create(ctx.Context(), thread.ThreadID, claims.UserID, pollInput); err != nil {
			_ = c.mediaService.DeleteByThread(ctx.Context(), thread.ThreadID)
			_ = c.threadService.Delete(ctx.Context(), thread.ThreadID, claims.UserID)
			return ctrlshared.HandleServiceError(ctx, err)
		}
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "thread created successfully"})
}

func (c *ThreadController) CreateEventThread(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	content := ctx.FormValue("content")
	files := ThreadMediaFiles(ctx)
	if err := threadmediamodel.ValidateThreadMediaCount(len(files)); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	var parentID *uint
	if raw := ctx.FormValue("parent_id"); raw != "" {
		if pid, err := strconv.ParseUint(raw, 10, 64); err == nil {
			v := uint(pid)
			parentID = &v
		}
	}

	eventID, _, err := ParseOptionalEventID(ctx)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	if eventID == nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "event_id is required"})
	}

	isOfficial := parseBoolForm(ctx, "is_official")

	var replyPermission *string
	if raw := strings.TrimSpace(ctx.FormValue("reply_permission")); raw != "" {
		replyPermission = &raw
	}

	pollInput, _, hasPoll, err := threadpollctrl.ParseCreatePollFromMultipart(ctx)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	if hasPoll && parentID != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "polls are only allowed on post-type threads"})
	}

	thread, err := c.threadService.CreateEventThread(ctx.Context(), claims.UserID, content, parentID, eventID, replyPermission, isOfficial)
	if err != nil {
		switch err.Error() {
		case "not allowed to reply to this thread", "forbidden", "blocked":
			return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: err.Error()})
		}
		return ctrlshared.HandleServiceError(ctx, err)
	}

	if len(files) > 0 {
		if _, err := c.mediaService.Upload(ctx.Context(), thread.ThreadID, claims.UserID, files); err != nil {
			_ = c.threadService.Delete(ctx.Context(), thread.ThreadID, claims.UserID)
			return ctrlshared.HandleServiceError(ctx, err)
		}
	}

	if hasPoll {
		if _, err := c.pollService.Create(ctx.Context(), thread.ThreadID, claims.UserID, pollInput); err != nil {
			_ = c.mediaService.DeleteByThread(ctx.Context(), thread.ThreadID)
			_ = c.threadService.Delete(ctx.Context(), thread.ThreadID, claims.UserID)
			return ctrlshared.HandleServiceError(ctx, err)
		}
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "thread created successfully"})
}

func parseBoolForm(ctx *fiber.Ctx, key string) bool {
	raw := strings.TrimSpace(ctx.FormValue(key))
	switch strings.ToLower(raw) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}

func (c *ThreadController) GetRoots(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	page, limit := ctrlshared.ParsePagination(ctx)

	threads, parents, total, err := c.threadService.GetForYouFeed(ctx.Context(), claims.UserID, page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}

	return ctx.Status(fiber.StatusOK).JSON(threaddto.PaginatedFeedResponse{
		Data: c.BuildUserFeedItemsNoPin(threads, parents, &claims.UserID),
		Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

// RefreshFeed drops the caller's For You feed ordering seed so the next
// GET /threads/ returns a reshuffled feed. Intended to be called from the
// frontend pull-to-refresh gesture before re-fetching page 1.
func (c *ThreadController) RefreshFeed(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	c.threadService.RefreshForYouFeed(ctx.Context(), claims.UserID)
	return ctx.SendStatus(fiber.StatusNoContent)
}

// GetNewForYouSince returns visible feed-eligible threads created after the
// supplied since_id, newest-first. Powers "Show N new posts" — the frontend
// calls this on banner click and prepends the results onto the existing feed
// instead of triggering /feed/refresh (which reshuffles the seed and can
// surface older threads, re-triggering the banner).
func (c *ThreadController) GetNewForYouSince(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	sinceID, _ := strconv.ParseUint(ctx.Query("since_id", "0"), 10, 64)
	limit, _ := strconv.Atoi(ctx.Query("limit", "50"))
	threads, parents, err := c.threadService.GetNewForYouSinceID(ctx.Context(), claims.UserID, uint(sinceID), limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}
	return ctx.Status(fiber.StatusOK).JSON(threaddto.PaginatedFeedResponse{
		Data: c.BuildUserFeedItemsNoPin(threads, parents, &claims.UserID),
		Meta: threaddto.FeedMeta{Total: int64(len(threads)), Page: 1, Limit: limit},
	})
}

// GetNewFeedCount returns the count of feed-eligible threads created after
// the supplied since_id, capped at 1000. Powers the "Show N new posts" banner.
func (c *ThreadController) GetNewFeedCount(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	sinceID, _ := strconv.ParseUint(ctx.Query("since_id", "0"), 10, 64)
	count, err := c.threadService.CountNewForYouSinceID(ctx.Context(), claims.UserID, uint(sinceID))
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}
	return ctx.JSON(fiber.Map{"count": count})
}

func (c *ThreadController) GetByID(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ctrlshared.ParseIDParam(ctx, "id", "thread")
	if !ok {
		return nil
	}

	thread, err := c.threadService.GetByID(id)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: "thread not found"})
	}

	if thread.CommunityID != nil && thread.Community != nil && thread.Community.Visibility == communitymodel.CommunityVisibilityPrivate {
		if !c.memberService.IsMember(*thread.CommunityID, claims.UserID) {
			return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: "forbidden"})
		}
	}
	if err := c.threadService.CanViewUserThreads(ctx.Context(), claims.UserID, thread.UserID); err != nil {
		if err.Error() == "this account is private" {
			return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: err.Error()})
		}
		if err.Error() == "blocked" {
			return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: "you can't view this account"})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}

	// Hide a muted author's post on direct GET (one-way mute), matching feed-level filtering.
	if muted, mErr := c.threadService.IsAuthorMutedBy(claims.UserID, thread.UserID); mErr == nil && muted {
		return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: "thread not found"})
	}

	media, _ := c.mediaService.GetByThreadID(id)
	fb := c.fbNoPin(&claims.UserID)
	fb.LoadAvatarsForUserIDs(ThreadUserIDs([]threadmodel.Thread{*thread}))
	fb.LoadPollsForThreads([]threadmodel.Thread{*thread})
	return ctx.Status(fiber.StatusOK).JSON(fb.ToFeedItem(thread, media))
}

func (c *ThreadController) GetChain(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ctrlshared.ParseIDParam(ctx, "id", "thread")
	if !ok {
		return nil
	}

	root, err := c.threadService.GetByID(id)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: "thread not found"})
	}

	if root.CommunityID != nil && root.Community != nil && root.Community.Visibility == communitymodel.CommunityVisibilityPrivate {
		if !c.memberService.IsMember(*root.CommunityID, claims.UserID) {
			return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: "forbidden"})
		}
	}

	page, limit := ctrlshared.ParsePagination(ctx)

	threads, total, err := c.threadService.GetChain(ctx.Context(), id, claims.UserID, page, limit)
	if err != nil {
		if err.Error() == "this account is private" {
			return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	// Build a set of thread IDs already present in the chain so we don't
	// re-fetch parents that are part of the chain itself.
	chainIDs := make(map[uint]struct{}, len(threads))
	for _, t := range threads {
		chainIDs[t.ThreadID] = struct{}{}
	}

	// Collect unique parent IDs that are not already part of the chain, then
	// fetch them all in one batched call instead of one query per parent.
	parentCache := make(map[uint]*threadmodel.Thread)
	missingParentIDs := make([]uint, 0)
	for _, t := range threads {
		if t.ParentID == nil {
			continue
		}
		pid := *t.ParentID
		if _, inChain := chainIDs[pid]; inChain {
			continue
		}
		if _, queued := parentCache[pid]; queued {
			continue
		}
		parentCache[pid] = nil // mark queued (dedups repeated parents)
		missingParentIDs = append(missingParentIDs, pid)
	}
	if len(missingParentIDs) > 0 {
		if fetched, err := c.threadService.GetByIDs(missingParentIDs); err == nil {
			for i := range fetched {
				p := fetched[i]
				parentCache[p.ThreadID] = &p
			}
		}
	}

	// Build a parallel parents slice (nil where no external parent exists).
	parents := make([]*threadmodel.Thread, len(threads))
	for i, t := range threads {
		if t.ParentID != nil {
			parents[i] = parentCache[*t.ParentID]
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(threaddto.PaginatedFeedResponse{
		Data: c.BuildUserFeedItemsNoPin(threads, parents, &claims.UserID),
		Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *ThreadController) Update(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ctrlshared.ParseIDParam(ctx, "id", "thread")
	if !ok {
		return nil
	}

	if threadpollctrl.HasPollFieldsInForm(ctx) {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "poll cannot be edited"})
	}

	content := ctx.FormValue("content")

	if _, err := c.threadService.Update(ctx.Context(), id, claims.UserID, content); err != nil {
		if err.Error() == "forbidden" {
			return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: "forbidden"})
		}
		return ctrlshared.HandleServiceError(ctx, err)
	}

	keepIDs, keepSent, err := parseKeepMediaIDs(ctx)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	files := ThreadMediaFiles(ctx)
	if err := threadmediamodel.ValidateThreadMediaCount(len(keepIDs) + len(files)); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(authdto.ErrorResponse{Message: err.Error()})
	}
	if keepSent || len(files) > 0 {
		if err := c.mediaService.UpdateGranular(ctx.Context(), id, claims.UserID, keepIDs, files); err != nil {
			return ctrlshared.HandleServiceError(ctx, err)
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "thread updated successfully"})
}

func (c *ThreadController) UpdateReplyPermission(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ctrlshared.ParseIDParam(ctx, "id", "thread")
	if !ok {
		return nil
	}

	var req threaddto.UpdateReplyPermissionRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid request body"})
	}

	if err := c.threadService.UpdateReplyPermission(ctx.Context(), id, claims.UserID, req.ReplyPermission); err != nil {
		switch err.Error() {
		case "forbidden":
			return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: "forbidden"})
		case "thread not found":
			return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
		default:
			return ctrlshared.HandleServiceError(ctx, err)
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "reply permission updated successfully"})
}

func (c *ThreadController) Delete(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ctrlshared.ParseIDParam(ctx, "id", "thread")
	if !ok {
		return nil
	}

	if err := c.threadService.Delete(ctx.Context(), id, claims.UserID); err != nil {
		if err.Error() == "forbidden" {
			return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: "forbidden"})
		}
		return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	_ = c.mediaService.DeleteByThread(ctx.Context(), id)
	_ = c.pollService.DeleteByThread(ctx.Context(), id)

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "thread deleted successfully"})
}

func (c *ThreadController) GetFollowingFeed(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	page, limit := ctrlshared.ParsePagination(ctx)

	threads, parents, total, err := c.threadService.GetFollowingFeed(claims.UserID, page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}

	return ctx.Status(fiber.StatusOK).JSON(threaddto.PaginatedFeedResponse{
		Data: c.BuildUserFeedItemsNoPin(threads, parents, &claims.UserID),
		Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *ThreadController) GetMyPosts(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	id, ok := ctrlshared.ParseIDParam(ctx, "userID", "user")
	if !ok {
		return nil
	}
	targetUserID := id
	if err := c.threadService.CanViewUserThreads(ctx.Context(), claims.UserID, targetUserID); err != nil {
		if err.Error() == "this account is private" {
			return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: err.Error()})
		}
		if err.Error() == "blocked" {
			return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: "you can't view this account"})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}
	page, limit := ctrlshared.ParsePagination(ctx)

	threads, parents, total, err := c.threadService.GetMyPosts(ctx.Context(), claims.UserID, targetUserID, page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}

	return ctx.Status(fiber.StatusOK).JSON(threaddto.PaginatedFeedResponse{
		Data: c.BuildUserFeedItemsWithPin(threads, parents, &claims.UserID),
		Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *ThreadController) GetMyReplies(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	id, ok := ctrlshared.ParseIDParam(ctx, "userID", "user")
	if !ok {
		return nil
	}
	targetUserID := id
	if err := c.threadService.CanViewUserThreads(ctx.Context(), claims.UserID, targetUserID); err != nil {
		if err.Error() == "this account is private" {
			return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: err.Error()})
		}
		if err.Error() == "blocked" {
			return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: "you can't view this account"})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}
	page, limit := ctrlshared.ParsePagination(ctx)

	threads, parents, total, err := c.threadService.GetMyReplies(ctx.Context(), claims.UserID, targetUserID, page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}

	return ctx.Status(fiber.StatusOK).JSON(threaddto.PaginatedFeedResponse{
		Data: c.BuildUserFeedItemsWithPin(threads, parents, &claims.UserID),
		Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *ThreadController) GetMyMediaThread(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	id, ok := ctrlshared.ParseIDParam(ctx, "userId", "user")
	if !ok {
		return nil
	}
	targetUserID := id
	if err := c.threadService.CanViewUserThreads(ctx.Context(), claims.UserID, targetUserID); err != nil {
		if err.Error() == "this account is private" {
			return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: err.Error()})
		}
		if err.Error() == "blocked" {
			return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: "you can't view this account"})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}
	page, limit := ctrlshared.ParsePagination(ctx)

	threads, _, total, err := c.threadService.GetMyMediaThread(ctx.Context(), claims.UserID, targetUserID, page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}

	mediaByThread := ctrlshared.LoadThreadMediaByThread(c.mediaService, threads)
	data := make([]threaddto.MediaOnlyItem, len(threads))
	for i, thread := range threads {
		media := make([]threadmediadto.ThreadMediaResponse, len(mediaByThread[thread.ThreadID]))
		for j, item := range mediaByThread[thread.ThreadID] {
			media[j] = threadmediadto.ThreadMediaResponse{
				MediaID: item.MediaID,
				URL:     c.mediaService.GetURL(item.ObjectName),
			}
		}
		data[i] = threaddto.MediaOnlyItem{
			ThreadID: thread.ThreadID,
			Media:    media,
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(threaddto.PaginatedMediaOnlyResponse{
		Data: data,
		Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func ThreadUserIDs(threads []threadmodel.Thread) []uint {
	ids := make([]uint, 0, len(threads))
	for _, t := range threads {
		if t.Type == threadmodel.ThreadTypeRepost && t.OriginalThread != nil {
			ids = append(ids, t.OriginalThread.User.UserID, t.UserID)
		} else {
			ids = append(ids, t.User.UserID)
		}
	}
	return ids
}

func ThreadMediaFiles(ctx *fiber.Ctx) []*multipart.FileHeader {
	form, err := ctx.MultipartForm()
	if err != nil || form == nil {
		return nil
	}
	return form.File["files"]
}

func parseKeepMediaIDs(ctx *fiber.Ctx) ([]uint, bool, error) {
	var values []string
	sent := false

	form, err := ctx.MultipartForm()
	if err == nil && form != nil {
		if raw, ok := form.Value["keep_media_ids"]; ok {
			values = append(values, raw...)
			sent = true
		}
	}

	if !sent {
		if raw := ctx.FormValue("keep_media_ids"); raw != "" {
			values = append(values, raw)
			sent = true
		}
	}

	if !sent {
		return nil, false, nil
	}

	ids := make([]uint, 0, len(values))
	seen := make(map[uint]struct{}, len(values))
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			id, err := strconv.ParseUint(part, 10, 64)
			if err != nil || id == 0 {
				return nil, true, errInvalidKeepMediaID()
			}
			mediaID := uint(id)
			if _, ok := seen[mediaID]; ok {
				continue
			}
			seen[mediaID] = struct{}{}
			ids = append(ids, mediaID)
		}
	}

	return ids, true, nil
}

func errInvalidKeepMediaID() error {
	return fiber.NewError(fiber.StatusBadRequest, "invalid keep_media_ids")
}

func ParseOptionalEventID(ctx *fiber.Ctx) (*string, bool, error) {
	raw, ok := formValueWithPresence(ctx, "event_id")
	if !ok {
		raw, ok = formValueWithPresence(ctx, "eventId")
	}
	if !ok {
		return nil, false, nil
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, true, nil
	}
	return &raw, true, nil
}

func formValueWithPresence(ctx *fiber.Ctx, key string) (string, bool) {
	if form, err := ctx.MultipartForm(); err == nil && form != nil {
		if raw, ok := form.Value[key]; ok {
			if len(raw) == 0 {
				return "", true
			}
			return raw[0], true
		}
	}

	body := ctx.Request().PostArgs()
	if body.Has(key) {
		return string(body.Peek(key)), true
	}
	return "", false
}
