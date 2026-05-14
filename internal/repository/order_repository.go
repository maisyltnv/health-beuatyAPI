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
