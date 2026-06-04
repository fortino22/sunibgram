package foxsessionctrl

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	foxsessionmodel "github.com/sunibgram/domain/model/fox_session"
	foxsessionsvciface "github.com/sunibgram/domain/service/fox_session"
	threadmediasvciface "github.com/sunibgram/domain/service/thread_media"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	dtoutils "github.com/sunibgram/webapi/dto/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
	foxsessiondto "github.com/sunibgram/webapi/dto/fox_session"
	threaddto "github.com/sunibgram/webapi/dto/thread"
	threadmediadto "github.com/sunibgram/webapi/dto/thread_media"
)

type FoxSessionController struct {
	sessionService foxsessionsvciface.IFoxSessionService
	mediaService   threadmediasvciface.IThreadMediaService
}

func NewFoxSessionController(
	sessionService foxsessionsvciface.IFoxSessionService,
	mediaService threadmediasvciface.IThreadMediaService,
) *FoxSessionController {
	return &FoxSessionController{sessionService: sessionService, mediaService: mediaService}
}

// List returns the authenticated user's Fox sessions (chats + summaries),
// newest first, one row per session.
func (c *FoxSessionController) List(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	page, limit := ctrlshared.ParsePagination(ctx)
	q := strings.TrimSpace(ctx.Query("q"))

	sessions, total, err := c.sessionService.ListSessions(ctx.Context(), claims.UserID, page, limit, q)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}

	// Collect summary-session thread ids for current media (detail view).
	ids := make([]uint, 0, len(sessions))
	for _, s := range sessions {
		if s.Session.Type == foxsessionmodel.TypeSummary && s.Session.ThreadID != nil {
			ids = append(ids, *s.Session.ThreadID)
		}
	}
	mediaByThread := ctrlshared.LoadThreadMediaByID(c.mediaService, ids)

	data := make([]foxsessiondto.SessionListItem, len(sessions))
	for i, s := range sessions {
		item := foxsessiondto.SessionListItem{
			ID:        s.Session.ID,
			Type:      s.Session.Type,
			Title:     s.Session.Title,
			IsStale:   s.Session.IsStale,
			UpdatedAt: dtoutils.Wrap(s.Session.UpdatedAt),
			Media:     []threadmediadto.ThreadMediaResponse{},
		}
		if s.Session.Type == foxsessionmodel.TypeChat {
			item.MessageCount = s.TurnCount
			if s.Entry != nil {
				item.Preview = s.Entry.Body
			}
		} else {
			var threadID uint
			if s.Session.ThreadID != nil {
				threadID = *s.Session.ThreadID
			}
			item.ThreadID = threadID
			if s.Entry != nil {
				item.Content = s.Entry.Snapshot
				item.Summary = s.Entry.Body
			}
			media := mediaByThread[threadID]
			mediaResp := make([]threadmediadto.ThreadMediaResponse, len(media))
			for j, m := range media {
				mediaResp[j] = threadmediadto.ThreadMediaResponse{MediaID: m.MediaID, URL: c.mediaService.GetURL(m.ObjectName)}
			}
			item.Media = mediaResp
		}
		data[i] = item
	}

	return ctx.Status(fiber.StatusOK).JSON(foxsessiondto.PaginatedSessionsResponse{
		Data: data,
		Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

// Detail returns a chat session's turns so the client can re-open and continue it.
func (c *FoxSessionController) Detail(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	sessionID, ok := ctrlshared.ParseIDParam(ctx, "id", "session")
	if !ok {
		return nil
	}

	turns, err := c.sessionService.GetSessionTurns(ctx.Context(), claims.UserID, sessionID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}

	resp := foxsessiondto.SessionDetailResponse{ID: sessionID, Turns: make([]foxsessiondto.SessionTurn, len(turns))}
	for i, t := range turns {
		resp.Turns[i] = foxsessiondto.SessionTurn{Question: t.Question, Answer: t.Answer}
	}
	return ctx.Status(fiber.StatusOK).JSON(resp)
}
