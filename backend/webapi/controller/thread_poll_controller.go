package controller

import (
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
)

type ThreadPollController struct {
	pollService service.IThreadPollService
}

func NewThreadPollController(pollService service.IThreadPollService) *ThreadPollController {
	return &ThreadPollController{pollService: pollService}
}

func (c *ThreadPollController) Vote(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}
	threadID, ok := ParseIDParam(ctx, "id", "thread")
	if !ok {
		return nil
	}
	var req dto.VotePollRequest
	if err := ctx.BodyParser(&req); err != nil || req.OptionID == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "option_id is required"})
	}
	_, _, err := c.pollService.Vote(ctx.Context(), threadID, req.OptionID, claims.UserID)
	if err != nil {
		return HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "vote cast successfully"})
}

// ParseCreatePollFromMultipart extracts optional poll fields from a multipart create-thread form.
// Returns (input, hasPoll, error). hasPoll is true when the caller intends to create a poll.
func ParseCreatePollFromMultipart(ctx *fiber.Ctx) (service.CreatePollInput, []*multipart.FileHeader, bool, error) {
	form, err := ctx.MultipartForm()
	if err != nil || form == nil {
		return service.CreatePollInput{}, nil, false, nil
	}
	labels := pollOptionLabels(form)
	durationRaw := firstFormValue(form, "poll_duration_seconds")
	question := strings.TrimSpace(firstFormValue(form, "poll_question"))

	if len(labels) == 0 && durationRaw == "" && question == "" {
		return service.CreatePollInput{}, nil, false, nil
	}

	durationSecs, err := strconv.Atoi(strings.TrimSpace(durationRaw))
	if err != nil {
		return service.CreatePollInput{}, nil, true, fiber.NewError(fiber.StatusBadRequest, "poll_duration_seconds is required and must be an integer")
	}
	images := form.File["poll_option_images"]
	input := service.CreatePollInput{
		Question:        question,
		Labels:          labels,
		Images:          images,
		DurationSeconds: durationSecs,
	}
	return input, images, true, nil
}

func pollOptionLabels(form *multipart.Form) []string {
	if values, ok := form.Value["poll_options"]; ok && len(values) > 0 {
		// support both repeated form fields and a single comma-separated string
		if len(values) == 1 && strings.Contains(values[0], ",") {
			parts := strings.Split(values[0], ",")
			out := make([]string, 0, len(parts))
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					out = append(out, p)
				}
			}
			return out
		}
		out := make([]string, 0, len(values))
		for _, v := range values {
			v = strings.TrimSpace(v)
			if v != "" {
				out = append(out, v)
			}
		}
		return out
	}
	return nil
}

func firstFormValue(form *multipart.Form, key string) string {
	if values, ok := form.Value[key]; ok && len(values) > 0 {
		return values[0]
	}
	return ""
}

// HasPollFieldsInForm reports whether the multipart form contains any poll-related field.
// Used by Update to reject poll-edits.
func HasPollFieldsInForm(ctx *fiber.Ctx) bool {
	form, err := ctx.MultipartForm()
	if err != nil || form == nil {
		return false
	}
	keys := []string{"poll_question", "poll_options", "poll_duration_seconds"}
	for _, k := range keys {
		if v, ok := form.Value[k]; ok && len(v) > 0 {
			return true
		}
	}
	if files, ok := form.File["poll_option_images"]; ok && len(files) > 0 {
		return true
	}
	return false
}
