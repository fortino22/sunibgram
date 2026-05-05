package service

import "github.com/sunibgram/domain/model"

type CreateOrderResult struct {
	Order      *model.Order
	PaymentURL string
	Token      string
}

type IOrderService interface {
	CreateOrder(userID uint, amount int64, purpose, itemID, itemName string) (*CreateOrderResult, error)
	BuyPremium(userID uint) (*CreateOrderResult, error)
	GetByID(id uint) (*model.Order, error)
	GetByOrderID(orderID string) (*model.Order, error)
	GetByUserID(userID uint, page, limit int) ([]model.Order, int64, error)
	GetAll(page, limit int) ([]model.Order, int64, error)
	HandleNotification(payload map[string]interface{}) error
	CancelOrder(id uint) error
}
