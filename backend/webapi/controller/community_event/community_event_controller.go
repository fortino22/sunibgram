package communityeventctrl

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	communityeventmodel "github.com/sunibgram/domain/model/community_event"
	usermodel "github.com/sunibgram/domain/model/user"
	userprofilemodel "github.com/sunibgram/domain/model/user_profile"
	communityeventsvciface "github.com/sunibgram/domain/service/community_event"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	utils "github.com/sunibgram/webapi/dto/_shared"
	communityeventdto "github.com/sunibgram/webapi/dto/community_event"
	threaddto "github.com/sunibgram/webapi/dto/thread"
)

type CommunityEventController struct {
	svc communityeventsvciface.ICommunityEventService
}

func NewCommunityEventController(svc communityeventsvciface.ICommunityEventService) *CommunityEventController {
	return &CommunityEventController{svc: svc}
}

// ---- helpers ----

func parseFormTime(raw string) (time.Time, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return time.Time{}, errors.New("time is required")
	}
	layouts := []string{"2006-01-02 15:04:05", "2006-01-02 15:04", time.RFC3339Nano, time.RFC3339}
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, errors.New("invalid time format")
}

func parseInviteeEmails(ctx *fiber.Ctx) []string {
	out := make([]string, 0)
	if form, err := ctx.MultipartForm(); err == nil && form != nil {
		for _, v := range form.Value["invitee_emails"] {
			for _, part := range strings.Split(v, ",") {
				if e := strings.TrimSpace(part); e != "" {
					out = append(out, e)
				}
			}
		}
	}
	return out
}

func parseInviteeUserIDs(ctx *fiber.Ctx) []uint {
	out := make([]uint, 0)
	if form, err := ctx.MultipartForm(); err == nil && form != nil {
		for _, v := range form.Value["invitee_user_ids"] {
			for _, part := range strings.Split(v, ",") {
				if id, pErr := strconv.ParseUint(strings.TrimSpace(part), 10, 64); pErr == nil && id > 0 {
					out = append(out, uint(id))
				}
			}
		}
	}
	return out
}

func (c *CommunityEventController) authorMini(user usermodel.User, profile *userprofilemodel.UserProfile) communityeventdto.AuthorMini {
	avatar := ""
	if profile != nil {
		avatar = ctrlshared.BuildMediaURL(profile.AvatarObjectName, c.svc.GetURL)
	}
	return communityeventdto.AuthorMini{
		UserID:    user.UserID,
		Username:  user.UserName,
		Name:      user.Name,
		AvatarURL: avatar,
	}
}

func (c *CommunityEventController) toCommentResponse(comment *communityeventmodel.CommunityEventComment, pinnedID *uint) communityeventdto.CommentResponse {
	isPinned := pinnedID != nil && *pinnedID == comment.ID
	return communityeventdto.CommentResponse{
		ID:        comment.ID,
		Author:    c.authorMini(comment.User, comment.Profile),
		Content:   comment.Content,
		IsPinned:  isPinned,
		CreatedAt: utils.FormattedTime{Time: comment.CreatedAt},
	}
}

func (c *CommunityEventController) toDetailResponse(detail *communityeventsvciface.CommunityEventDetail) communityeventdto.CommunityEventResponse {
	event := detail.Event

	invitees := make([]communityeventdto.AuthorMini, 0, len(detail.Invitees))
	for _, inv := range detail.Invitees {
		invitees = append(invitees, c.authorMini(inv.User, inv.Profile))
	}

	reactions := make([]communityeventdto.ReactionSummaryItem, 0, len(detail.ReactionCounts))
	for _, rc := range detail.ReactionCounts {
		reactions = append(reactions, communityeventdto.ReactionSummaryItem{Emoji: rc.Emoji, Count: rc.Count})
	}

	var pinned *communityeventdto.CommentResponse
	if detail.PinnedComment != nil {
		resp := c.toCommentResponse(detail.PinnedComment, event.PinnedCommentID)
		pinned = &resp
	}

	return communityeventdto.CommunityEventResponse{
		ID:              event.ID,
		CommunityID:     event.CommunityID,
		CreatedBy:       c.authorMini(event.Creator, nil),
		Title:           event.Title,
		Description:     event.Description,
		Location:        event.Location,
		StartTime:       communityeventdto.EventTime{Time: event.StartTime},
		EndTime:         communityeventdto.EventTime{Time: event.EndTime},
		BannerURL:       ctrlshared.BuildMediaURL(event.BannerObjectName, c.svc.GetURL),
		Invitees:        invitees,
		ReactionSummary: reactions,
		MyReaction:      detail.MyReaction,
		PinnedComment:   pinned,
		CommentCount:    event.CommentCount,
		CanComment:      detail.CanComment,
		IsAdmin:         detail.IsAdmin,
		CreatedAt:       utils.FormattedTime{Time: event.CreatedAt},
		UpdatedAt:       utils.FormattedTime{Time: event.UpdatedAt},
	}
}

func (c *CommunityEventController) toListResponse(event communityeventmodel.CommunityEvent) communityeventdto.CommunityEventListResponse {
	return communityeventdto.CommunityEventListResponse{
		ID:           event.ID,
		CommunityID:  event.CommunityID,
		Title:        event.Title,
		Description:  event.Description,
		Location:     event.Location,
		StartTime:    communityeventdto.EventTime{Time: event.StartTime},
		EndTime:      communityeventdto.EventTime{Time: event.EndTime},
		BannerURL:    ctrlshared.BuildMediaURL(event.BannerObjectName, c.svc.GetURL),
		CommentCount: event.CommentCount,
		CreatedAt:    utils.FormattedTime{Time: event.CreatedAt},
	}
}

// ---- event handlers ----

func (c *CommunityEventController) Create(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	communityID, ok := ctrlshared.ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}

	start, err := parseFormTime(ctx.FormValue("start_time"))
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, errors.New("start_time is invalid"))
	}
	end, err := parseFormTime(ctx.FormValue("end_time"))
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, errors.New("end_time is invalid"))
	}

	input := communityeventsvciface.CreateCommunityEventInput{
		Title:          ctx.FormValue("title"),
		Description:    ctx.FormValue("description"),
		Location:       ctx.FormValue("location"),
		StartTime:      start,
		EndTime:        end,
		InviteeEmails:  parseInviteeEmails(ctx),
		InviteeUserIDs: parseInviteeUserIDs(ctx),
	}
	if banner, fErr := ctx.FormFile("banner"); fErr == nil {
		input.Banner = banner
	}

	event, err := c.svc.Create(ctx.Context(), communityID, claims.UserID, input)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	detail, err := c.svc.GetByID(ctx.Context(), event.ID, claims.UserID)
	if err != nil {
		return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "event created successfully", "id": event.ID})
	}
	return ctx.Status(fiber.StatusCreated).JSON(c.toDetailResponse(detail))
}

func (c *CommunityEventController) SearchInviteeCandidates(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	communityID, ok := ctrlshared.ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}
	candidates, err := c.svc.SearchInviteeCandidates(ctx.Context(), communityID, claims.UserID, ctx.Query("q"))
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	data := make([]communityeventdto.AuthorMini, 0, len(candidates))
	for _, cand := range candidates {
		data = append(data, communityeventdto.AuthorMini{
			UserID:    cand.UserID,
			Username:  cand.Username,
			Name:      cand.Name,
			AvatarURL: ctrlshared.BuildMediaURL(cand.AvatarObjectName, c.svc.GetURL),
		})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": data})
}

func (c *CommunityEventController) GetByCommunity(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	communityID, ok := ctrlshared.ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}
	page, limit := ctrlshared.ParsePagination(ctx)

	events, total, err := c.svc.GetByCommunity(ctx.Context(), communityID, claims.UserID, page, limit)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	data := make([]communityeventdto.CommunityEventListResponse, 0, len(events))
	for _, e := range events {
		data = append(data, c.toListResponse(e))
	}
	return ctx.Status(fiber.StatusOK).JSON(communityeventdto.PaginatedCommunityEventResponse{
		Data: data,
		Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *CommunityEventController) GetByID(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	eventID, ok := ctrlshared.ParseIDParam(ctx, "eventId", "event")
	if !ok {
		return nil
	}
	detail, err := c.svc.GetByID(ctx.Context(), eventID, claims.UserID)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(c.toDetailResponse(detail))
}

func (c *CommunityEventController) Update(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	eventID, ok := ctrlshared.ParseIDParam(ctx, "eventId", "event")
	if !ok {
		return nil
	}

	start, err := parseFormTime(ctx.FormValue("start_time"))
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, errors.New("start_time is invalid"))
	}
	end, err := parseFormTime(ctx.FormValue("end_time"))
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, errors.New("end_time is invalid"))
	}

	input := communityeventsvciface.CreateCommunityEventInput{
		Title:       ctx.FormValue("title"),
		Description: ctx.FormValue("description"),
		Location:    ctx.FormValue("location"),
		StartTime:   start,
		EndTime:     end,
		ClearBanner: strings.EqualFold(ctx.FormValue("clear_banner"), "true"),
	}
	if emails := parseInviteeEmails(ctx); len(emails) > 0 {
		input.InviteeEmails = emails
	}
	if ids := parseInviteeUserIDs(ctx); len(ids) > 0 {
		input.InviteeUserIDs = ids
	}
	if banner, fErr := ctx.FormFile("banner"); fErr == nil {
		input.Banner = banner
	}

	event, err := c.svc.Update(ctx.Context(), eventID, claims.UserID, input)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	detail, err := c.svc.GetByID(ctx.Context(), event.ID, claims.UserID)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(c.toDetailResponse(detail))
}

func (c *CommunityEventController) Delete(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	eventID, ok := ctrlshared.ParseIDParam(ctx, "eventId", "event")
	if !ok {
		return nil
	}
	if err := c.svc.Delete(ctx.Context(), eventID, claims.UserID); err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "event deleted successfully"})
}

// ---- comment handlers ----

func (c *CommunityEventController) GetComments(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	eventID, ok := ctrlshared.ParseIDParam(ctx, "eventId", "event")
	if !ok {
		return nil
	}
	page, limit := ctrlshared.ParsePagination(ctx)

	list, err := c.svc.GetComments(ctx.Context(), eventID, claims.UserID, page, limit)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	data := make([]communityeventdto.CommentResponse, 0, len(list.Comments))
	for i := range list.Comments {
		data = append(data, c.toCommentResponse(&list.Comments[i], list.PinnedCommentID))
	}
	return ctx.Status(fiber.StatusOK).JSON(communityeventdto.PaginatedCommentResponse{
		Data:            data,
		PinnedCommentID: list.PinnedCommentID,
		IsAdmin:         list.IsAdmin,
		Meta:            threaddto.FeedMeta{Total: list.Total, Page: page, Limit: limit},
	})
}

func (c *CommunityEventController) AddComment(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	eventID, ok := ctrlshared.ParseIDParam(ctx, "eventId", "event")
	if !ok {
		return nil
	}
	var req communityeventdto.AddCommentRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctrlshared.HandleServiceError(ctx, errors.New("invalid request body"))
	}
	comment, err := c.svc.AddComment(ctx.Context(), eventID, claims.UserID, req.Content)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusCreated).JSON(c.toCommentResponse(comment, nil))
}

func (c *CommunityEventController) UpdateComment(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	eventID, ok := ctrlshared.ParseIDParam(ctx, "eventId", "event")
	if !ok {
		return nil
	}
	commentID, ok := ctrlshared.ParseIDParam(ctx, "commentId", "comment")
	if !ok {
		return nil
	}
	var req communityeventdto.UpdateCommentRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctrlshared.HandleServiceError(ctx, errors.New("invalid request body"))
	}
	comment, err := c.svc.UpdateComment(ctx.Context(), eventID, commentID, claims.UserID, req.Content)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(c.toCommentResponse(comment, nil))
}

func (c *CommunityEventController) DeleteComment(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	eventID, ok := ctrlshared.ParseIDParam(ctx, "eventId", "event")
	if !ok {
		return nil
	}
	commentID, ok := ctrlshared.ParseIDParam(ctx, "commentId", "comment")
	if !ok {
		return nil
	}
	if err := c.svc.DeleteComment(ctx.Context(), eventID, commentID, claims.UserID); err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "comment deleted successfully"})
}

func (c *CommunityEventController) PinComment(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	eventID, ok := ctrlshared.ParseIDParam(ctx, "eventId", "event")
	if !ok {
		return nil
	}
	commentID, ok := ctrlshared.ParseIDParam(ctx, "commentId", "comment")
	if !ok {
		return nil
	}
	if err := c.svc.PinComment(ctx.Context(), eventID, commentID, claims.UserID); err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "comment pinned successfully"})
}

func (c *CommunityEventController) UnpinComment(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	eventID, ok := ctrlshared.ParseIDParam(ctx, "eventId", "event")
	if !ok {
		return nil
	}
	if err := c.svc.UnpinComment(ctx.Context(), eventID, claims.UserID); err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "comment unpinned successfully"})
}

// ---- reaction handlers ----

func (c *CommunityEventController) SetReaction(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	eventID, ok := ctrlshared.ParseIDParam(ctx, "eventId", "event")
	if !ok {
		return nil
	}
	var req communityeventdto.SetReactionRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctrlshared.HandleServiceError(ctx, errors.New("invalid request body"))
	}
	emoji, err := c.svc.SetReaction(ctx.Context(), eventID, claims.UserID, req.Emoji)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"my_reaction": emoji})
}

func (c *CommunityEventController) RemoveReaction(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	eventID, ok := ctrlshared.ParseIDParam(ctx, "eventId", "event")
	if !ok {
		return nil
	}
	if err := c.svc.RemoveReaction(ctx.Context(), eventID, claims.UserID); err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "reaction removed successfully"})
}
