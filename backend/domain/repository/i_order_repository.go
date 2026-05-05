package repository

import "github.com/sunibgram/domain/model"

type IOrderRepository interface {
	Save(order *model.Order) error
	FindByID(id uint) (*model.Order, error)
	FindByOrderID(orderID string) (*model.Order, error)
	FindByUserID(userID uint, page, limit int) ([]model.Order, int64, error)
	FindAll(page, limit int) ([]model.Order, int64, error)
	UpdateStatus(orderID, status, paymentType string) error
	Delete(id uint) error
}
