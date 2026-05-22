package repository

import (
	"errors"

	"gorm.io/gorm"

	ordermodel "github.com/sunibgram/domain/model/order"
	orderrepoiface "github.com/sunibgram/domain/repository/order"
)

type orderRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewOrderRepository(writeDB, readDB *gorm.DB) orderrepoiface.IOrderRepository {
	return &orderRepository{writeDB: writeDB, readDB: readDB}
}

func (r *orderRepository) Save(order *ordermodel.Order) error {
	return r.writeDB.Create(order).Error
}

func (r *orderRepository) FindByID(id uint) (*ordermodel.Order, error) {
	var order ordermodel.Order
	if err := r.readDB.First(&order, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) FindByOrderID(orderID string) (*ordermodel.Order, error) {
	var order ordermodel.Order
	if err := r.readDB.First(&order, "order_id = ?", orderID).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) FindPendingByUserIDAndPurpose(userID uint, purpose string) (*ordermodel.Order, error) {
	var order ordermodel.Order
	err := r.readDB.Where("user_id = ? AND purpose = ? AND status = 'PENDING'", userID, purpose).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) FindByUserID(userID uint, page, limit int) ([]ordermodel.Order, int64, error) {
	var orders []ordermodel.Order
	var total int64
	r.readDB.Model(&ordermodel.Order{}).Where("user_id = ?", userID).Count(&total)
	offset := (page - 1) * limit
	if err := r.readDB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&orders).Error; err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

func (r *orderRepository) FindAll(page, limit int) ([]ordermodel.Order, int64, error) {
	var orders []ordermodel.Order
	var total int64
	r.readDB.Model(&ordermodel.Order{}).Count(&total)
	offset := (page - 1) * limit
	if err := r.readDB.Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&orders).Error; err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

func (r *orderRepository) UpdateStatus(orderID, status, paymentType string) error {
	result := r.writeDB.Model(&ordermodel.Order{}).
		Where("order_id = ?", orderID).
		Updates(map[string]interface{}{
			"status":       status,
			"payment_type": paymentType,
		})
	if result.RowsAffected == 0 {
		return errors.New("order not found")
	}
	return result.Error
}

func (r *orderRepository) Delete(id uint) error {
	result := r.writeDB.Delete(&ordermodel.Order{}, "id = ?", id)
	if result.RowsAffected == 0 {
		return errors.New("order not found")
	}
	return result.Error
}
