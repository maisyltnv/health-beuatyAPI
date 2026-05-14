package repository

import (
	"context"

	"shopapi/internal/model"

	"gorm.io/gorm"
)

// ProductRepository persists products.
type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(ctx context.Context, p *model.Product) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *ProductRepository) GetByID(ctx context.Context, id uint64) (*model.Product, error) {
	var p model.Product
	if err := r.db.WithContext(ctx).First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepository) List(ctx context.Context, limit, offset int) ([]model.Product, int64, error) {
	var items []model.Product
	var total int64
	q := r.db.WithContext(ctx).Model(&model.Product{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *ProductRepository) Update(ctx context.Context, p *model.Product) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *ProductRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Product{}, id).Error
}
