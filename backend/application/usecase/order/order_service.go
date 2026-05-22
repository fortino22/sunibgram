package orderuc

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"

	"github.com/sunibgram/application/helper"
	ordermodel "github.com/sunibgram/domain/model/order"
	premiumsubscriptionmodel "github.com/sunibgram/domain/model/premium_subscription"
	cacherepoiface "github.com/sunibgram/domain/repository/cache"
	communityrepoiface "github.com/sunibgram/domain/repository/community"
	orderrepoiface "github.com/sunibgram/domain/repository/order"
	premiumsubscriptionrepoiface "github.com/sunibgram/domain/repository/premium_subscription"
	userrepoiface "github.com/sunibgram/domain/repository/user"
	ordersvciface "github.com/sunibgram/domain/service/order"
)

const PremiumAmount int64 = 10000

type OrderService struct {
	orderRepo               orderrepoiface.IOrderRepository
	userRepo                userrepoiface.IUserRepository
	premiumSubscriptionRepo premiumsubscriptionrepoiface.IPremiumSubscriptionRepository
	communityRepo           communityrepoiface.ICommunityRepository
	cacheRepo               cacherepoiface.ICacheRepository
	snapClient              snap.Client
}

func NewOrderService(orderRepo orderrepoiface.IOrderRepository, userRepo userrepoiface.IUserRepository, premiumSubscriptionRepo premiumsubscriptionrepoiface.IPremiumSubscriptionRepository, communityRepo communityrepoiface.ICommunityRepository, cacheRepo cacherepoiface.ICacheRepository) ordersvciface.IOrderService {
	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")
	envStr := os.Getenv("MIDTRANS_ENV")

	env := midtrans.Sandbox
	if envStr == "production" {
		env = midtrans.Production
	}

	var client snap.Client
	client.New(serverKey, env)

	return &OrderService{
		orderRepo:               orderRepo,
		userRepo:                userRepo,
		premiumSubscriptionRepo: premiumSubscriptionRepo,
		communityRepo:           communityRepo,
		cacheRepo:               cacheRepo,
		snapClient:              client,
	}
}

func (s *OrderService) CreateOrder(userID uint, amount int64, purpose, itemID, itemName string) (*ordersvciface.CreateOrderResult, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}

	orderID := fmt.Sprintf("ORDER-%s-%d", uuid.NewString()[:8], time.Now().UnixMilli())

	req := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  orderID,
			GrossAmt: amount,
		},
		Items: &[]midtrans.ItemDetails{
			{
				ID:    itemID,
				Name:  itemName,
				Price: amount,
				Qty:   1,
			},
		},
	}

	snapResp, err := s.snapClient.CreateTransaction(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create midtrans transaction: %w", err)
	}

	order := &ordermodel.Order{
		OrderID: orderID,
		UserID:  userID,
		Amount:  amount,
		Purpose: purpose,
		Status:  "PENDING",
	}
	if err := s.orderRepo.Save(order); err != nil {
		return nil, err
	}

	return &ordersvciface.CreateOrderResult{
		Order:      order,
		PaymentURL: snapResp.RedirectURL,
		Token:      snapResp.Token,
	}, nil
}

func (s *OrderService) BuyPremium(userID uint) (*ordersvciface.CreateOrderResult, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if user.IsPremium {
		return nil, errors.New("user already has premium")
	}

	if _, err := s.orderRepo.FindPendingByUserIDAndPurpose(userID, "PREMIUM"); err == nil {
		return nil, errors.New("pending premium order already exists")
	}

	return s.CreateOrder(userID, PremiumAmount, "PREMIUM", "SUNIBGRAM-PREMIUM", "Sunibgram Premium")
}

func (s *OrderService) GetByID(id uint) (*ordermodel.Order, error) {
	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("order not found")
	}
	return order, nil
}

func (s *OrderService) GetByOrderID(orderID string) (*ordermodel.Order, error) {
	order, err := s.orderRepo.FindByOrderID(orderID)
	if err != nil {
		return nil, errors.New("order not found")
	}
	return order, nil
}

func (s *OrderService) GetByUserID(userID uint, page, limit int) ([]ordermodel.Order, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 10)
	return s.orderRepo.FindByUserID(userID, page, limit)
}

func (s *OrderService) GetAll(page, limit int) ([]ordermodel.Order, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 10)
	return s.orderRepo.FindAll(page, limit)
}

func (s *OrderService) HandleNotification(payload map[string]interface{}) error {
	orderID, _ := payload["order_id"].(string)
	if orderID == "" {
		return errors.New("missing order_id in notification")
	}

	transactionStatus, _ := payload["transaction_status"].(string)
	fraudStatus, _ := payload["fraud_status"].(string)
	paymentType, _ := payload["payment_type"].(string)

	status := resolveStatus(transactionStatus, fraudStatus)
	if status == "" {
		return nil
	}

	order, err := s.orderRepo.FindByOrderID(orderID)
	if err != nil {
		return errors.New("order not found")
	}

	if err := s.orderRepo.UpdateStatus(orderID, status, paymentType); err != nil {
		return err
	}

	if status == "PAID" && order.Purpose == "PREMIUM" {
		if err := s.userRepo.UpdatePremium(order.UserID, true); err != nil {
			return fmt.Errorf("failed to set user premium: %w", err)
		}
		now := time.Now()
		subscription := &premiumsubscriptionmodel.PremiumSubscription{
			UserID:                order.UserID,
			IsTrial:               false,
			SubscriptionStartDate: now,
			SubscriptionEndDate:   now.AddDate(0, 1, 0),
		}
		if err := s.premiumSubscriptionRepo.Upsert(subscription); err != nil {
			return fmt.Errorf("failed to create subscription: %w", err)
		}
		if err := s.communityRepo.ReactivateByOwnerID(order.UserID); err != nil {
			return fmt.Errorf("failed to reactivate communities: %w", err)
		}
		if s.cacheRepo != nil {
			ctx := context.Background()
			_ = s.cacheRepo.InvalidateCommunityRequestsGlobal(ctx)
			_ = s.cacheRepo.InvalidateCommunityThreadsGlobal(ctx)
			_ = s.cacheRepo.InvalidateThreadAllCommunitiesGlobal(ctx)
		}
	}

	return nil
}

func (s *OrderService) CancelOrder(id uint) error {
	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		return errors.New("order not found")
	}
	if order.Status != "PENDING" {
		return errors.New("only pending orders can be cancelled")
	}
	return s.orderRepo.UpdateStatus(order.OrderID, "CANCELLED", order.PaymentType)
}

func resolveStatus(transactionStatus, fraudStatus string) string {
	switch transactionStatus {
	case "capture":
		if fraudStatus == "accept" {
			return "PAID"
		}
		return "CANCELLED"
	case "settlement":
		return "PAID"
	case "cancel", "deny", "expire":
		return "CANCELLED"
	case "pending":
		return "PENDING"
	default:
		return ""
	}
}
