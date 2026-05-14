package repository

import (
	"context"

	"shopapi/internal/model"

	"gorm.io/gorm"
)

// OrderRepository persists orders (scaffold for order placement in a later iteration).
type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, o *model.Order) error {
	return r.db.WithContext(ctx).Create(o).Error
}

// ListByUserID returns orders for a user, newest first.
func (r *OrderRepository) ListByUserID(ctx context.Context, userID uint64, limit, offset int) ([]model.Order, int64, error) {
	var items []model.Order
	var total int64
	q := r.db.WithContext(ctx).Model(&model.Order{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&items).Error
	return items, total, err
}
