package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	"github.com/sunibgram/dto/utils"
)

type ReportController struct {
	reportService service.IReportService
}

func NewReportController(reportService service.IReportService) *ReportController {
	return &ReportController{reportService: reportService}
}

func (c *ReportController) Create(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	var req dto.CreateReportRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}
	violationID, err := uuid.Parse(req.ViolationID)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid violation id"})
	}
	if err := c.reportService.Create(claims.UserID, req.ReportedUserID, req.ReportedPostID, violationID, req.Explanation); err != nil {
		return HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "report submitted successfully"})
}

func (c *ReportController) GetAll(ctx *fiber.Ctx) error {
	reports, err := c.reportService.GetAll()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": toReportResponses(reports)})
}

func (c *ReportController) GetByID(ctx *fiber.Ctx) error {
	id, ok := ParseUUIDParam(ctx, "id", "report")
	if !ok {
		return nil
	}
	report, err := c.reportService.GetByID(id)
	if err != nil {
		return HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(toReportResponse(report))
}

func (c *ReportController) UpdateStatus(ctx *fiber.Ctx) error {
	id, ok := ParseUUIDParam(ctx, "id", "report")
	if !ok {
		return nil
	}

	var req dto.UpdateReportStatusRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}
	if err := c.reportService.UpdateStatus(id, req.Status); err != nil {
		return HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "report status updated successfully"})
}

func (c *ReportController) Delete(ctx *fiber.Ctx) error {
	id, ok := ParseUUIDParam(ctx, "id", "report")
	if !ok {
		return nil
	}
	if err := c.reportService.Delete(id); err != nil {
		return HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "report deleted successfully"})
}

func toReportResponse(detail *model.ReportDetail) dto.ReportResponse {
	report := &detail.Report
	response := dto.ReportResponse{
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
		violation := toViolationResponse(&report.Violation)
		response.Violation = &violation
	}
	response.ReporterUser = toReportUserResponse(detail.ReporterUser)
	response.ReportedUser = toReportUserResponse(detail.ReportedUser)
	return response
}

func toReportResponses(reports []model.ReportDetail) []dto.ReportResponse {
	responses := make([]dto.ReportResponse, len(reports))
	for i := range reports {
		responses[i] = toReportResponse(&reports[i])
	}
	return responses
}

func toReportUserResponse(user *model.ReportUserInfo) *dto.ReportUserResponse {
	if user == nil {
		return nil
	}
	response := &dto.ReportUserResponse{
		UserID:     user.UserID,
		Username:   user.Username,
		UserAvatar: user.UserAvatar,
		IsPremium:  user.IsPremium,
	}
	if user.Faculty != nil {
		response.Faculty = &dto.ReportFacultyResponse{
			FacultyID:   user.Faculty.FacultyID,
			FacultyName: user.Faculty.FacultyName,
		}
	}
	return response
}
