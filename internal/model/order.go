package model

import "time"

// OrderStatus is persisted as a string on the order row.
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusCancelled OrderStatus = "cancelled"
)

// Order is a customer purchase in LAK.
type Order struct {
	ID                uint64     `gorm:"primaryKey" json:"id"`
	UserID            uint64     `gorm:"index;not null" json:"user_id"`
	TotalAmountLAK    float64    `gorm:"type:decimal(18,4);not null" json:"total_amount_lak"`
	Status            OrderStatus `gorm:"type:varchar(32);not null;default:pending" json:"status"`
	PaymentReceiptURL string     `gorm:"size:2048" json:"payment_receipt_url"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func (Order) TableName() string {
	return "orders"
}
