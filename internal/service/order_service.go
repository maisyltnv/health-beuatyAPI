package service

import (
	"context"
	"errors"
	"math"
	"sort"

	"shopapi/internal/model"
	"shopapi/internal/repository"

	"gorm.io/gorm"
)

const maxOrderQty = 9999

// OrderService handles checkout: orders are built from product lines with server-side pricing.
type OrderService struct {
	orders   *repository.OrderRepository
	products *repository.ProductRepository
}

func NewOrderService(orders *repository.OrderRepository, products *repository.ProductRepository) *OrderService {
	return &OrderService{orders: orders, products: products}
}

type OrderLineInput struct {
	ProductID uint64
	Quantity  int
}

type PlaceOrderInput struct {
	UserID            uint64
	Lines             []OrderLineInput
	PaymentReceiptURL string
}

func roundMoneyLAK(x float64) float64 {
	return math.Round(x*100) / 100
}

// Place creates an order from product lines; total is computed from current product prices (no client total).
func (s *OrderService) Place(ctx context.Context, in PlaceOrderInput) (*model.Order, error) {
	if len(in.Lines) == 0 {
		return nil, errors.New("order must have at least one line item")
	}
	merged := make(map[uint64]int)
	for _, line := range in.Lines {
		if line.ProductID == 0 {
			return nil, errors.New("invalid product_id")
		}
		if line.Quantity < 1 || line.Quantity > maxOrderQty {
			return nil, errors.New("invalid quantity")
		}
		merged[line.ProductID] += line.Quantity
		if merged[line.ProductID] > maxOrderQty {
			return nil, errors.New("quantity per product exceeds limit")
		}
	}

	var items []model.OrderItem
	var total float64
	pids := make([]uint64, 0, len(merged))
	for pid := range merged {
		pids = append(pids, pid)
	}
	sort.Slice(pids, func(i, j int) bool { return pids[i] < pids[j] })
	for _, pid := range pids {
		qty := merged[pid]
		p, err := s.products.GetByID(ctx, pid)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("product not found")
			}
			return nil, err
		}
		unit := roundMoneyLAK(p.FinalPriceLAK)
		lineTotal := roundMoneyLAK(unit * float64(qty))
		items = append(items, model.OrderItem{
			ProductID:    pid,
			ProductName:  p.Name,
			UnitPriceLAK: unit,
			Quantity:     qty,
			LineTotalLAK: lineTotal,
		})
		total += lineTotal
	}
	total = roundMoneyLAK(total)

	o := &model.Order{
		UserID:            in.UserID,
		TotalAmountLAK:    total,
		Status:            model.OrderStatusPending,
		PaymentReceiptURL: in.PaymentReceiptURL,
	}
	if err := s.orders.CreateWithItems(ctx, o, items); err != nil {
		return nil, err
	}
	return s.orders.GetByUser(ctx, o.ID, in.UserID)
}

// GetMine returns a single order for the user with items and products.
func (s *OrderService) GetMine(ctx context.Context, userID, orderID uint64) (*model.Order, error) {
	return s.orders.GetByUser(ctx, orderID, userID)
}

// ListMine returns paginated orders for the given user (includes line items).
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
