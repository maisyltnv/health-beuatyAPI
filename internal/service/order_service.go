package service

import (
	"context"

	"shopapi/internal/model"
	"shopapi/internal/repository"
)

// OrderService will back order placement endpoints; kept minimal in this scaffold phase.
type OrderService struct {
	orders *repository.OrderRepository
}

func NewOrderService(orders *repository.OrderRepository) *OrderService {
	return &OrderService{orders: orders}
}

type PlaceOrderInput struct {
	UserID            uint64
	TotalAmountLAK    float64
	PaymentReceiptURL string
}

func (s *OrderService) Place(ctx context.Context, in PlaceOrderInput) (*model.Order, error) {
	o := &model.Order{
		UserID:            in.UserID,
		TotalAmountLAK:    in.TotalAmountLAK,
		Status:            model.OrderStatusPending,
		PaymentReceiptURL: in.PaymentReceiptURL,
	}
	if err := s.orders.Create(ctx, o); err != nil {
		return nil, err
	}
	return o, nil
}

// ListMine returns paginated orders for the given user.
func (s *OrderService) ListMine(ctx context.Context, userID uint64, limit, offset int) ([]model.Order, int64, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	return s.orders.ListByUserID(ctx, userID, limit, offset)
}
