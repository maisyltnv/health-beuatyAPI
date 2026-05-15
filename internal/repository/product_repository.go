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
	if err := r.db.WithContext(ctx).Preload("Category").First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

// List returns products with category preloaded. If categoryID is non-nil, only that category.
func (r *ProductRepository) List(ctx context.Context, limit, offset int, categoryID *uint64) ([]model.Product, int64, error) {
	var items []model.Product
	var total int64
	q := r.db.WithContext(ctx).Model(&model.Product{})
	if categoryID != nil {
		q = q.Where("category_id = ?", *categoryID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Preload("Category").Order("id DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// CountByCategoryID counts products assigned to a category (direct assignment only).
func (r *ProductRepository) CountByCategoryID(ctx context.Context, categoryID uint64) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.Product{}).Where("category_id = ?", categoryID).Count(&n).Error
	return n, err
}

func (r *ProductRepository) Update(ctx context.Context, p *model.Product) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *ProductRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Product{}, id).Error
}
