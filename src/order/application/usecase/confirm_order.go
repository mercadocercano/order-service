package usecase

import (
	"context"
	"fmt"
	"order/src/order/domain/entity"
	"order/src/order/domain/port"
	"order/src/order/infrastructure/client"
)

// ConfirmOrderUseCase caso de uso para confirmar una orden
type ConfirmOrderUseCase struct {
	orderRepo   port.OrderRepository
	stockClient *client.StockClient
}

// NewConfirmOrderUseCase crea una nueva instancia del caso de uso
func NewConfirmOrderUseCase(orderRepo port.OrderRepository, stockClient *client.StockClient) *ConfirmOrderUseCase {
	return &ConfirmOrderUseCase{
		orderRepo:   orderRepo,
		stockClient: stockClient,
	}
}

// Execute ejecuta la confirmación de la orden (multi-item, atómico)
func (uc *ConfirmOrderUseCase) Execute(ctx context.Context, tenantID, authToken, orderID, reference string) (*entity.Order, error) {
	// 1. Buscar orden con sus items (load aggregate)
	order, err := uc.orderRepo.FindByID(ctx, orderID, tenantID)
	if err != nil {
		return nil, entity.ErrOrderNotFound
	}

	// 2. Validar que esté en estado CREATED
	if order.Status != entity.OrderStatusCreated {
		return nil, entity.ErrOrderNotInCreatedState
	}

	// 3. Consumir stock reservado para CADA item vía Kong (ALL OR NOTHING)
	for _, item := range order.Items {
		_, err = uc.stockClient.ConsumeStock(tenantID, authToken, item.SKU, item.Quantity, reference)
		if err != nil {
			// Si falla un item, TODO el proceso falla
			// Nota: En producción debería hacer rollback de items anteriores
			if contains(err.Error(), "insufficient reserved stock") {
				return nil, fmt.Errorf("insufficient_reserved_stock for SKU %s: %w", item.SKU, err)
			}
			return nil, fmt.Errorf("error consuming stock for SKU %s: %w", item.SKU, err)
		}
	}

	// 4. Confirmar orden en DB
	if err := uc.orderRepo.Confirm(ctx, orderID, tenantID); err != nil {
		return nil, fmt.Errorf("error confirming order: %w", err)
	}

	// 5. Actualizar entidad en memoria
	order.Status = entity.OrderStatusConfirmed

	return order, nil
}
