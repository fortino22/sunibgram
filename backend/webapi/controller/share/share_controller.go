package sharectrl

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	storageuc "github.com/sunibgram/application/usecase/storage"
	sharesvciface "github.com/sunibgram/domain/service/share"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
	sharedto "github.com/sunibgram/webapi/dto/share"
)

type ShareController struct {
	shareService   sharesvciface.IShareService
	storageUsecase *storageuc.StorageUsecase
}

func NewShareController(shareService sharesvciface.IShareService, storageUsecase *storageuc.StorageUsecase) *ShareController {
	return &ShareController{shareService: shareService, storageUsecase: storageUsecase}
}

func (c *ShareController) RecordCopy(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	threadID, ok := ctrlshared.ParseIDParam(ctx, "id", "thread")
	if !ok {
		return nil
	}
	count, err := c.shareService.RecordCopy(ctx.Context(), claims.UserID, threadID)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(sharedto.CopyShareResponse{
		Message:    "share recorded successfully",
		ShareCount: count,
	})
}

func (c *ShareController) ShareViaChat(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	threadID, ok := ctrlshared.ParseIDParam(ctx, "id", "thread")
	if !ok {
		return nil
	}
	var req sharedto.ShareViaChatRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid request body"})
	}
	sent, err := c.shareService.SendToUsers(ctx.Context(), claims.UserID, threadID, req.UserIDs, req.Message)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(sharedto.ChatShareResponse{
		Message:        "shared via chat successfully",
		RecipientCount: sent,
	})
}

func (c *ShareController) SearchUsers(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	query := strings.TrimSpace(ctx.Query("q"))
	results, err := c.shareService.SearchUsers(ctx.Context(), query, claims.UserID)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	data := make([]sharedto.ShareUserResponse, len(results))
	for i, u := range results {
		var avatar *string
		if c.storageUsecase != nil && u.AvatarObjectName != "" {
			url := c.storageUsecase.GetURL(u.AvatarObjectName)
			avatar = &url
		}
		data[i] = sharedto.ShareUserResponse{
			UserID:    u.UserID,
			Name:      u.Name,
			Username:  u.Username,
			IsPremium: u.IsPremium,
			AvatarURL: avatar,
		}
	}
	return ctx.Status(fiber.StatusOK).JSON(sharedto.SearchShareUsersResponse{Data: data})
}
