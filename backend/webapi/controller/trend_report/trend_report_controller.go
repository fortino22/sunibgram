package trendreportctrl

import (
	"github.com/gofiber/fiber/v2"

	trendreportmodel "github.com/sunibgram/domain/model/trend_report"
	trendreportsvciface "github.com/sunibgram/domain/service/trend_report"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
	trendreportdto "github.com/sunibgram/webapi/dto/trend_report"
)

type TrendReportController struct {
	service trendreportsvciface.ITrendReportService
}

func NewTrendReportController(service trendreportsvciface.ITrendReportService) *TrendReportController {
	return &TrendReportController{service: service}
}

func (c *TrendReportController) Create(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	var req trendreportdto.CreateTrendReportRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid request body"})
	}

	if err := c.service.CreateReport(ctx.Context(), claims.UserID, req.HashtagName, req.ReasonCode); err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "trend reported successfully"})
}

func (c *TrendReportController) ListReasons(ctx *fiber.Ctx) error {
	reasons, err := c.service.ListReasons(ctx.Context())
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": toReasonResponses(reasons)})
}

func toReasonResponses(reasons []trendreportmodel.TrendReportReason) []trendreportdto.TrendReportReasonResponse {
	out := make([]trendreportdto.TrendReportReasonResponse, len(reasons))
	for i, r := range reasons {
		out[i] = trendreportdto.TrendReportReasonResponse{Code: r.Code, Label: r.Label}
	}
	return out
}
