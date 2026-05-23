package feedbackctrl

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	feedbackmodel "github.com/sunibgram/domain/model/feedback"
	feedbacksvciface "github.com/sunibgram/domain/service/feedback"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	utils "github.com/sunibgram/webapi/dto/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
	feedbackdto "github.com/sunibgram/webapi/dto/feedback"
	webapiutils "github.com/sunibgram/webapi/utils"
)

type FeedbackController struct {
	svc feedbacksvciface.IFeedbackService
}

func NewFeedbackController(svc feedbacksvciface.IFeedbackService) *FeedbackController {
	return &FeedbackController{svc: svc}
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
	f, err := c.svc.Create(claims.UserID, req.Type, req.Title, req.Body)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "feedback created successfully",
		"data":    toResponse(f, feedbackmodel.VoteCounts{}, 0),
	})
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
	data := make([]feedbackdto.FeedbackResponse, len(views))
	for i := range views {
		data[i] = toResponse(&views[i].Feedback, views[i].Counts, views[i].MyVote)
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
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": toResponse(&view.Feedback, view.Counts, view.MyVote),
	})
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
	return c.vote(ctx, 1)
}

func (c *FeedbackController) Dislike(ctx *fiber.Ctx) error {
	return c.vote(ctx, -1)
}

func (c *FeedbackController) ClearVote(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	id, ok := parseUUIDParam(ctx, "id", "feedback")
	if !ok {
		return nil
	}
	counts, err := c.svc.ClearVote(claims.UserID, id)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(feedbackdto.VoteResponse{
		Likes: counts.Likes, Dislikes: counts.Dislikes, MyVote: 0,
	})
}

func (c *FeedbackController) vote(ctx *fiber.Ctx, value int) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	id, ok := parseUUIDParam(ctx, "id", "feedback")
	if !ok {
		return nil
	}
	counts, err := c.svc.Vote(claims.UserID, id, value)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(feedbackdto.VoteResponse{
		Likes: counts.Likes, Dislikes: counts.Dislikes, MyVote: value,
	})
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

func toResponse(f *feedbackmodel.Feedback, counts feedbackmodel.VoteCounts, myVote int) feedbackdto.FeedbackResponse {
	return feedbackdto.FeedbackResponse{
		ID:     f.ID.String(),
		Type:   f.Type,
		Title:  f.Title,
		Body:   f.Body,
		Status: f.Status,
		Author: feedbackdto.FeedbackAuthor{
			UserID:   f.User.UserID,
			Username: f.User.UserName,
			Name:     f.User.Name,
		},
		Likes:     counts.Likes,
		Dislikes:  counts.Dislikes,
		MyVote:    myVote,
		CreatedAt: utils.Wrap(f.CreatedAt),
		UpdatedAt: utils.Wrap(f.UpdatedAt),
	}
}
