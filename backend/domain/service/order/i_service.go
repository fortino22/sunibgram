package ordersvciface

import ordermodel "github.com/sunibgram/domain/model/order"

type CreateOrderResult struct {
	Order      *ordermodel.Order
	PaymentURL string
	Token      string
}

type IOrderService interface {
	CreateOrder(userID uint, amount int64, purpose, itemID, itemName string) (*CreateOrderResult, error)
	BuyPremium(userID uint) (*CreateOrderResult, error)
	GetByID(id uint) (*ordermodel.Order, error)
	GetByOrderID(orderID string) (*ordermodel.Order, error)
	GetByUserID(userID uint, page, limit int) ([]ordermodel.Order, int64, error)
	GetAll(page, limit int) ([]ordermodel.Order, int64, error)
	HandleNotification(payload map[string]interface{}) error
	CancelOrder(id uint) error
}
