package repository

import ordermodel "github.com/sunibgram/domain/model/order"

type IOrderRepository interface {
	Save(order *ordermodel.Order) error
	FindByID(id uint) (*ordermodel.Order, error)
	FindByOrderID(orderID string) (*ordermodel.Order, error)
	FindPendingByUserIDAndPurpose(userID uint, purpose string) (*ordermodel.Order, error)
	FindByUserID(userID uint, page, limit int) ([]ordermodel.Order, int64, error)
	FindAll(page, limit int) ([]ordermodel.Order, int64, error)
	UpdateStatus(orderID, status, paymentType string) error
	Delete(id uint) error
}
