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

type shippingRequest struct {
	RecipientName string `json:"recipient_name" binding:"required"`
	Phone         string `json:"phone" binding:"required"`
	Province      string `json:"province" binding:"required"`
	AddressDetail string `json:"address_detail" binding:"required"`
}

type placeOrderRequest struct {
	Items             []orderLineRequest `json:"items" binding:"required,min=1,dive"`
	Shipping          shippingRequest    `json:"shipping" binding:"required"`
	PaymentMethod     string             `json:"payment_method" binding:"required,oneof=bcel_qr cod"`
	PaymentReceiptURL string             `json:"payment_receipt_url"`
}

// ListByPhone is a public endpoint for customers to track orders by shipping phone (no JWT).
// Newest orders first. Pagination: ?phone=...&page=1&limit=10
func (h *OrderHandler) ListByPhone(c *gin.Context) {
	phone := c.Query("phone")
	if phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "phone query parameter is required"})
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	res, err := h.orders.ListByPhone(c.Request.Context(), phone, page, limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *OrderHandler) ShippingConfig(c *gin.Context) {
	c.JSON(http.StatusOK, h.orders.ShippingConfig())
}

func (h *OrderHandler) QuoteShipping(c *gin.Context) {
	subtotal, err := strconv.ParseFloat(c.DefaultQuery("subtotal_lak", "0"), 64)
	if err != nil || subtotal < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subtotal_lak"})
		return
	}
	c.JSON(http.StatusOK, h.orders.QuoteShipping(subtotal))
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
		UserID: uid,
		Lines:  lines,
		Shipping: service.ShippingInput{
			RecipientName: req.Shipping.RecipientName,
			Phone:         req.Shipping.Phone,
			Province:      req.Shipping.Province,
			AddressDetail: req.Shipping.AddressDetail,
		},
		PaymentMethod:     req.PaymentMethod,
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
