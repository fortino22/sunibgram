package controller

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	"github.com/sunibgram/webapi/ctxutil"
)

type OrderController struct {
	orderService service.IOrderService
}

func NewOrderController(orderService service.IOrderService) *OrderController {
	return &OrderController{orderService: orderService}
}

func (c *OrderController) Create(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	var req dto.CreateOrderRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}

	purpose := req.Purpose
	if purpose == "" {
		purpose = "GENERAL"
	}
	itemID := req.ItemID
	if itemID == "" {
		itemID = "SUNIBGRAM-ORDER"
	}
	itemName := req.ItemName
	if itemName == "" {
		itemName = "Sunibgram Order"
	}

	result, err := c.orderService.CreateOrder(claims.UserID, req.Amount, purpose, itemID, itemName)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusCreated).JSON(dto.CreateOrderResponse{
		Order:      toOrderResponse(result.Order),
		PaymentURL: result.PaymentURL,
		Token:      result.Token,
	})
}

func (c *OrderController) BuyPremium(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	result, err := c.orderService.BuyPremium(claims.UserID)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusCreated).JSON(dto.CreateOrderResponse{
		Order:      toOrderResponse(result.Order),
		PaymentURL: result.PaymentURL,
		Token:      result.Token,
	})
}

func (c *OrderController) GetMyOrders(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))

	orders, total, err := c.orderService.GetByUserID(claims.UserID, page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	data := make([]dto.OrderResponse, len(orders))
	for i, o := range orders {
		data[i] = toOrderResponse(&o)
	}

	return ctx.JSON(dto.PaginatedOrderResponse{
		Data: data,
		Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *OrderController) GetAll(ctx *fiber.Ctx) error {
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))

	orders, total, err := c.orderService.GetAll(page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	data := make([]dto.OrderResponse, len(orders))
	for i, o := range orders {
		data[i] = toOrderResponse(&o)
	}

	return ctx.JSON(dto.PaginatedOrderResponse{
		Data: data,
		Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *OrderController) GetByID(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid order id"})
	}

	order, err := c.orderService.GetByID(uint(id))
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.JSON(toOrderResponse(order))
}

func (c *OrderController) Cancel(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid order id"})
	}

	order, err := c.orderService.GetByID(uint(id))
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	if order.UserID != claims.UserID && claims.RoleName != "admin" {
		return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: "forbidden"})
	}

	if err := c.orderService.CancelOrder(uint(id)); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.JSON(fiber.Map{"message": "order cancelled successfully"})
}

func (c *OrderController) MidtransNotification(ctx *fiber.Ctx) error {
	var payload map[string]interface{}
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid payload"})
	}

	if err := c.orderService.HandleNotification(payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.SendStatus(fiber.StatusOK)
}

func toOrderResponse(o *model.Order) dto.OrderResponse {
	return dto.OrderResponse{
		ID:          o.ID,
		OrderID:     o.OrderID,
		UserID:      o.UserID,
		Amount:      o.Amount,
		Purpose:     o.Purpose,
		Status:      o.Status,
		PaymentType: o.PaymentType,
		CreatedAt:   o.CreatedAt,
		UpdatedAt:   o.UpdatedAt,
	}
}
