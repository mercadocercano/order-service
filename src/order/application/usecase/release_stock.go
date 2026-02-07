package usecase

import (
	"fmt"
	"order/src/order/application/request"
	"order/src/order/application/response"
	"order/src/order/infrastructure/client"
)

// ReleaseStockUseCase caso de uso para liberar stock reservado
type ReleaseStockUseCase struct {
	stockClient *client.StockClient
}

// NewReleaseStockUseCase crea una nueva instancia del caso de uso
func NewReleaseStockUseCase(stockClient *client.StockClient) *ReleaseStockUseCase {
	return &ReleaseStockUseCase{
		stockClient: stockClient,
	}
}

// Execute ejecuta la liberación de stock
func (uc *ReleaseStockUseCase) Execute(tenantID, authToken string, req *request.ReleaseStockRequest) (*response.ReleaseStockResponse, error) {
	// Llamar a stock-service vía Kong
	stockResp, err := uc.stockClient.ReleaseStock(tenantID, authToken, req.SKU, req.Quantity, req.Reference)
	if err != nil {
		// Si es error de stock reservado insuficiente, propagarlo como 409
		if contains(err.Error(), "insufficient reserved stock") || contains(err.Error(), "409") {
			return nil, fmt.Errorf("insufficient_reserved_stock: %w", err)
		}
		return nil, fmt.Errorf("error releasing stock: %w", err)
	}

	// Mapear respuesta
	return &response.ReleaseStockResponse{
		Released:  true,
		SKU:       stockResp.SKU,
		Quantity:  stockResp.ReleasedQty,
		Reference: stockResp.Reference,
	}, nil
}
