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
