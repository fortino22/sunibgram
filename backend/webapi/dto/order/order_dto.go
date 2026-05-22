package orderdto

import (
	"time"

	threaddto "github.com/sunibgram/webapi/dto/thread"
)

type OrderResponse struct {
	ID          uint      `json:"id"`
	OrderID     string    `json:"order_id"`
	UserID      uint      `json:"user_id"`
	Amount      int64     `json:"amount"`
	Purpose     string    `json:"purpose"`
	Status      string    `json:"status"`
	PaymentType string    `json:"payment_type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateOrderResponse struct {
	Order      OrderResponse `json:"order"`
	PaymentURL string        `json:"payment_url"`
	Token      string        `json:"token"`
}

type PaginatedOrderResponse struct {
	Data []OrderResponse    `json:"data"`
	Meta threaddto.FeedMeta `json:"meta"`
}
