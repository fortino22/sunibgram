package feedbackctrl

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	feedbacksvciface "github.com/sunibgram/domain/service/feedback"
	userprofilesvciface "github.com/sunibgram/domain/service/user_profile"
	"github.com/sunibgram/infrastructure/cache"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	utils "github.com/sunibgram/webapi/dto/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
	feedbackdto "github.com/sunibgram/webapi/dto/feedback"
	webapiutils "github.com/sunibgram/webapi/utils"
)

type FeedbackController struct {
	svc        feedbacksvciface.IFeedbackService
	profileSvc userprofilesvciface.IUserProfileService
	cache      *cache.Cache
}

func NewFeedbackController(svc feedbacksvciface.IFeedbackService, profileSvc userprofilesvciface.IUserProfileService, c *cache.Cache) *FeedbackController {
	return &FeedbackController{svc: svc, profileSvc: profileSvc, cache: c}
}

func (c *FeedbackController) Create(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	var req feedbackdto.CreateFeedbackRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid request body"})
	}
	if err := c.svc.Create(claims.UserID, req.Type, req.Title, req.Body); err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "feedback created successfully"})
}

func (c *FeedbackController) List(ctx *fiber.Ctx) error {
	viewerID := optionalViewerID(ctx)
	fType := ctx.Query("type")
	status := ctx.Query("status")
	page, limit := webapiutils.ParsePagination(ctx)

	views, total, err := c.svc.List(viewerID, fType, status, page, limit)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}

	authorIDs := make([]uint, 0, len(views))
	for i := range views {
		authorIDs = append(authorIDs, views[i].Feedback.UserID)
	}
	fb := c.feedBuilder(viewerID)
	fb.LoadAvatarsForUserIDs(authorIDs)

	data := make([]feedbackdto.FeedbackResponse, len(views))
	for i := range views {
		data[i] = c.toResponse(&views[i], fb)
	}
	return ctx.Status(fiber.StatusOK).JSON(feedbackdto.PaginatedFeedbackResponse{
		Data: data,
		Meta: feedbackdto.FeedbackListMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *FeedbackController) GetByID(ctx *fiber.Ctx) error {
	id, ok := parseUUIDParam(ctx, "id", "feedback")
	if !ok {
		return nil
	}
	viewerID := optionalViewerID(ctx)
	view, err := c.svc.GetByID(viewerID, id)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	fb := c.feedBuilder(viewerID)
	fb.LoadAvatarsForUserIDs([]uint{view.Feedback.UserID})
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": c.toResponse(view, fb)})
}

func (c *FeedbackController) UpdateStatus(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	id, ok := parseUUIDParam(ctx, "id", "feedback")
	if !ok {
		return nil
	}
	var req feedbackdto.UpdateStatusRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid request body"})
	}
	if err := c.svc.UpdateStatus(claims.UserID, id, req.Status); err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "feedback status updated successfully"})
}

func (c *FeedbackController) Delete(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	id, ok := parseUUIDParam(ctx, "id", "feedback")
	if !ok {
		return nil
	}
	if err := c.svc.Delete(claims.UserID, claims.RoleName, id); err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "feedback deleted successfully"})
}

func (c *FeedbackController) Like(ctx *fiber.Ctx) error {
	return c.vote(ctx, 1, "feedback liked successfully")
}

func (c *FeedbackController) Dislike(ctx *fiber.Ctx) error {
	return c.vote(ctx, -1, "feedback disliked successfully")
}

func (c *FeedbackController) vote(ctx *fiber.Ctx, value int, successMsg string) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	id, ok := parseUUIDParam(ctx, "id", "feedback")
	if !ok {
		return nil
	}
	if err := c.svc.Vote(claims.UserID, id, value); err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": successMsg})
}

func (c *FeedbackController) feedBuilder(viewerID uint) webapiutils.FeedBuilder {
	var vid *uint
	if viewerID != 0 {
		v := viewerID
		vid = &v
	}
	return webapiutils.FeedBuilder{
		ProfileSvc: c.profileSvc,
		Cache:      c.cache,
		ViewerID:   vid,
	}
}

func (c *FeedbackController) toResponse(v *feedbacksvciface.FeedbackView, fb webapiutils.FeedBuilder) feedbackdto.FeedbackResponse {
	f := &v.Feedback
	overlaid := fb.AuthorField(f.User)
	author := feedbackdto.FeedbackAuthor{
		UserID:     overlaid.ID,
		Username:   overlaid.Username,
		Name:       overlaid.Name,
		UserAvatar: overlaid.AvatarURL,
		IsPremium:  overlaid.IsPremium,
	}
	return feedbackdto.FeedbackResponse{
		ID:        f.ID.String(),
		Type:      f.Type,
		Title:     f.Title,
		Body:      f.Body,
		Status:    f.Status,
		Author:    author,
		Likes:     v.Counts.Likes,
		Dislikes:  v.Counts.Dislikes,
		IsVote:    voteLabel(v.UserVote),
		CreatedAt: utils.Wrap(f.CreatedAt),
		UpdatedAt: utils.Wrap(f.UpdatedAt),
	}
}

func voteLabel(v int) *string {
	switch v {
	case 1:
		s := "like"
		return &s
	case -1:
		s := "dislike"
		return &s
	}
	return nil
}

func optionalViewerID(ctx *fiber.Ctx) uint {
	claims, ok := webapiutils.Get(ctx)
	if !ok {
		return 0
	}
	return claims.UserID
}

func parseUUIDParam(ctx *fiber.Ctx, paramName, entityName string) (uuid.UUID, bool) {
	id, err := uuid.Parse(ctx.Params(paramName))
	if err != nil {
		_ = ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid " + entityName + " id"})
		return uuid.Nil, false
	}
	return id, true
}
