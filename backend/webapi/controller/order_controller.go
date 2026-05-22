package controller

import (
	"crypto/sha512"
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"

	ordermodel "github.com/sunibgram/domain/model/order"
	ordersvciface "github.com/sunibgram/domain/service/order"
	authdto "github.com/sunibgram/webapi/dto/auth"
	orderdto "github.com/sunibgram/webapi/dto/order"
	threaddto "github.com/sunibgram/webapi/dto/thread"
)

type OrderController struct {
	orderService ordersvciface.IOrderService
}

func NewOrderController(orderService ordersvciface.IOrderService) *OrderController {
	return &OrderController{orderService: orderService}
}

func (c *OrderController) BuyPremium(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	result, err := c.orderService.BuyPremium(claims.UserID)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusCreated).JSON(orderdto.CreateOrderResponse{
		Order:      toOrderResponse(result.Order),
		PaymentURL: result.PaymentURL,
		Token:      result.Token,
	})
}

func (c *OrderController) GetMyOrders(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	page, limit := ParsePagination(ctx)

	orders, total, err := c.orderService.GetByUserID(claims.UserID, page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	data := make([]orderdto.OrderResponse, len(orders))
	for i, o := range orders {
		data[i] = toOrderResponse(&o)
	}

	return ctx.JSON(orderdto.PaginatedOrderResponse{
		Data: data,
		Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *OrderController) GetAll(ctx *fiber.Ctx) error {
	page, limit := ParsePagination(ctx)

	orders, total, err := c.orderService.GetAll(page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	data := make([]orderdto.OrderResponse, len(orders))
	for i, o := range orders {
		data[i] = toOrderResponse(&o)
	}

	return ctx.JSON(orderdto.PaginatedOrderResponse{
		Data: data,
		Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *OrderController) GetByID(ctx *fiber.Ctx) error {
	id, ok := ParseIDParam(ctx, "id", "order")
	if !ok {
		return nil
	}

	order, err := c.orderService.GetByID(id)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.JSON(toOrderResponse(order))
}

func (c *OrderController) Cancel(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ParseIDParam(ctx, "id", "order")
	if !ok {
		return nil
	}

	order, err := c.orderService.GetByID(id)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
	}
	if order.UserID != claims.UserID && claims.RoleName != "admin" {
		return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: "forbidden"})
	}

	if err := c.orderService.CancelOrder(id); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.JSON(fiber.Map{"message": "order cancelled successfully"})
}

func (c *OrderController) MidtransNotification(ctx *fiber.Ctx) error {
	var payload map[string]interface{}
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid payload"})
	}

	if err := verifyMidtransSignature(payload); err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(authdto.ErrorResponse{Message: "invalid signature"})
	}

	if err := c.orderService.HandleNotification(payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.SendStatus(fiber.StatusOK)
}

func verifyMidtransSignature(payload map[string]interface{}) error {
	orderID, _ := payload["order_id"].(string)
	statusCode, _ := payload["status_code"].(string)
	grossAmount, _ := payload["gross_amount"].(string)
	incomingSig, _ := payload["signature_key"].(string)

	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")
	raw := orderID + statusCode + grossAmount + serverKey
	hash := sha512.Sum512([]byte(raw))
	expected := fmt.Sprintf("%x", hash)

	if incomingSig != expected {
		return fmt.Errorf("signature mismatch")
	}
	return nil
}

func toOrderResponse(o *ordermodel.Order) orderdto.OrderResponse {
	return orderdto.OrderResponse{
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
