package response

// GetOrderResponse representa la respuesta de obtención de una orden
type GetOrderResponse struct {
	OrderID   string              `json:"order_id"`
	TenantID  string              `json:"tenant_id"`
	Status    string              `json:"status"`
	CreatedAt string              `json:"created_at"`
	Items     []OrderItemResponse `json:"items"`
}
