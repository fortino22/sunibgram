package helper

import (
	"mime/multipart"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/dto"
	"github.com/sunibgram/webapi/ctxutil"
)

func Error(ctx *fiber.Ctx, status int, message string) error {
	return ctx.Status(status).JSON(dto.ErrorResponse{Message: message})
}

func Unauthorized(ctx *fiber.Ctx) error {
	return Error(ctx, fiber.StatusUnauthorized, "unauthorized")
}

func Claims(ctx *fiber.Ctx) (ctxutil.Claims, bool) {
	return ctxutil.Get(ctx)
}

func ViewerID(ctx *fiber.Ctx) *uint {
	claims, ok := Claims(ctx)
	if !ok {
		return nil
	}
	return &claims.UserID
}

func ParamUint(ctx *fiber.Ctx, name string) (uint, error) {
	id, err := strconv.ParseUint(ctx.Params(name), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func OptionalFormUint(ctx *fiber.Ctx, name string) *uint {
	if raw := ctx.FormValue(name); raw != "" {
		if id, err := strconv.ParseUint(raw, 10, 64); err == nil {
			v := uint(id)
			return &v
		}
	}
	return nil
}

func PageLimit(ctx *fiber.Ctx) (int, int) {
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "20"))
	return page, limit
}

func MultipartFiles(ctx *fiber.Ctx, field string) []*multipart.FileHeader {
	form, err := ctx.MultipartForm()
	if err != nil || form == nil {
		return nil
	}
	return form.File[field]
}
