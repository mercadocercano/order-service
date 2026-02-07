package usecase

import (
	"context"
	"fmt"
	"order/src/order/application/request"
	"order/src/order/application/response"
	"order/src/order/domain/entity"
	"order/src/order/domain/port"
)

// CreateOrderUseCase caso de uso para crear una orden
type CreateOrderUseCase struct {
	orderRepo port.OrderRepository
}

// NewCreateOrderUseCase crea una nueva instancia del caso de uso
func NewCreateOrderUseCase(orderRepo port.OrderRepository) *CreateOrderUseCase {
	return &CreateOrderUseCase{
		orderRepo: orderRepo,
	}
}

// Execute ejecuta la creación de la orden (multi-item)
func (uc *CreateOrderUseCase) Execute(ctx context.Context, tenantID string, req *request.CreateOrderRequest) (*response.CreateOrderResponse, error) {
	// Construir items del aggregate
	var items []entity.OrderItem
	for _, itemReq := range req.Items {
		item, err := entity.NewOrderItem("", itemReq.SKU, itemReq.Quantity)
		if err != nil {
			return nil, fmt.Errorf("error creating order item: %w", err)
		}
		items = append(items, *item)
	}

	// Crear entidad Order (aggregate root)
	order, err := entity.NewOrder(tenantID, items)
	if err != nil {
		return nil, fmt.Errorf("error creating order entity: %w", err)
	}

	// Persistir orden con sus items (atomically)
	if err := uc.orderRepo.Save(ctx, order); err != nil {
		return nil, fmt.Errorf("error saving order: %w", err)
	}

	// Construir respuesta
	var itemsResp []response.CreateOrderItemResponse
	for _, item := range order.Items {
		itemsResp = append(itemsResp, response.CreateOrderItemResponse{
			ItemID:   item.ItemID,
			SKU:      item.SKU,
			Quantity: item.Quantity,
		})
	}

	return &response.CreateOrderResponse{
		OrderID:    order.OrderID,
		Items:      itemsResp,
		TotalItems: len(order.Items),
		Status:     string(order.Status),
	}, nil
}
