package repository

import (
	"context"

	"shopapi/internal/model"

	"gorm.io/gorm"
)

// OrderRepository persists orders and order line items.
type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// CreateWithItems creates an order and its line items in one transaction.
func (r *OrderRepository) CreateWithItems(ctx context.Context, o *model.Order, items []model.OrderItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(o).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].OrderID = o.ID
		}
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetByUser returns one order if it belongs to the user, with line items and product refs.
func (r *OrderRepository) GetByUser(ctx context.Context, orderID, userID uint64) (*model.Order, error) {
	var o model.Order
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", orderID, userID).
		Preload("Items").
		Preload("Items.Product").
		First(&o).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

// ListByUserID returns orders for a user, newest first, with line items (no deep product preload).
func (r *OrderRepository) ListByUserID(ctx context.Context, userID uint64, limit, offset int) ([]model.Order, int64, error) {
	var items []model.Order
	var total int64
	q := r.db.WithContext(ctx).Model(&model.Order{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Preload("Items").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&items).Error
	return items, total, err
}
