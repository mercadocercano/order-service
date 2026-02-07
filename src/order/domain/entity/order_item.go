package entity

import "github.com/google/uuid"

// OrderItem representa un item dentro de una orden (Entity dentro del Aggregate)
type OrderItem struct {
	ItemID   string `json:"item_id"`
	OrderID  string `json:"order_id"`
	SKU      string `json:"sku"`
	Quantity int    `json:"quantity"`
}

// NewOrderItem crea un nuevo item de orden
func NewOrderItem(orderID, sku string, quantity int) (*OrderItem, error) {
	if sku == "" {
		return nil, ErrSKURequired
	}
	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	return &OrderItem{
		ItemID:   uuid.New().String(),
		OrderID:  orderID,
		SKU:      sku,
		Quantity: quantity,
	}, nil
}
