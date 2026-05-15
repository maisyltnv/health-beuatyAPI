package handler

import (
	"errors"
	"net/http"
	"strconv"

	"shopapi/internal/middleware"
	"shopapi/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OrderHandler struct {
	orders *service.OrderService
}

func NewOrderHandler(orders *service.OrderService) *OrderHandler {
	return &OrderHandler{orders: orders}
}

type orderLineRequest struct {
	ProductID uint64 `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1,max=9999"`
}

type placeOrderRequest struct {
	Items               []orderLineRequest `json:"items" binding:"required,min=1,dive"`
	PaymentReceiptURL   string             `json:"payment_receipt_url"`
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
	lines := make([]service.OrderLineInput, 0, len(req.Items))
	for _, it := range req.Items {
		lines = append(lines, service.OrderLineInput{ProductID: it.ProductID, Quantity: it.Quantity})
	}
	o, err := h.orders.Place(c.Request.Context(), service.PlaceOrderInput{
		UserID:            uid,
		Lines:             lines,
		PaymentReceiptURL: req.PaymentReceiptURL,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, o)
}

// List returns the authenticated user's orders (newest first), each with line items.
func (h *OrderHandler) List(c *gin.Context) {
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
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	items, total, err := h.orders.ListMine(c.Request.Context(), uid, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total})
}

// Get returns one order for the authenticated user (with items and product snapshots).
func (h *OrderHandler) Get(c *gin.Context) {
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
	id, err := parseUintParam(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	o, err := h.orders.GetMine(c.Request.Context(), uid, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, o)
}
