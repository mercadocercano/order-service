package response

// OrderItemResponse representa un item en el listado
type OrderItemResponse struct {
	ItemID   string `json:"item_id"`
	OrderID  string `json:"order_id"`
	SKU      string `json:"sku"`
	Quantity int    `json:"quantity"`
}

// OrderListItem representa una orden en el listado
type OrderListItem struct {
	OrderID   string              `json:"order_id"`
	TenantID  string              `json:"tenant_id"`
	Status    string              `json:"status"`
	CreatedAt string              `json:"created_at"`
	Items     []OrderItemResponse `json:"items"`
}

// ListOrdersResponse representa la respuesta paginada de órdenes
type ListOrdersResponse struct {
	Items      []OrderListItem `json:"items"`
	TotalCount int             `json:"total_count"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}
