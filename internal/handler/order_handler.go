package handler

import (
	"net/http"

	"shopapi/internal/middleware"
	"shopapi/internal/service"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orders *service.OrderService
}

func NewOrderHandler(orders *service.OrderService) *OrderHandler {
	return &OrderHandler{orders: orders}
}

type placeOrderRequest struct {
	TotalAmountLAK    float64 `json:"total_amount_lak" binding:"required,gte=0"`
	PaymentReceiptURL string  `json:"payment_receipt_url"`
}

func (h *OrderHandler) Place(c *gin.Context) {
	uidVal, ok := c.Get(middleware.ContextUserIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	uid, ok := uidVal.(uint64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req placeOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	o, err := h.orders.Place(c.Request.Context(), service.PlaceOrderInput{
		UserID:            uid,
		TotalAmountLAK:    req.TotalAmountLAK,
		PaymentReceiptURL: req.PaymentReceiptURL,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, o)
}
