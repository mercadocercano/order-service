package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// StockAvailabilityResponse representa la respuesta de stock-service
type StockAvailabilityResponse struct {
	VariantSKU        string  `json:"variant_sku"`
	ProductSKU        string  `json:"product_sku"`
	AvailableQuantity float64 `json:"available_quantity"`
	ReservedQuantity  float64 `json:"reserved_quantity"`
	TotalQuantity     float64 `json:"total_quantity"`
	IsOutOfStock      bool    `json:"is_out_of_stock"`
	IsLowStock        bool    `json:"is_low_stock"`
}

// StockReserveRequest representa el request para reservar stock
type StockReserveRequest struct {
	SKU       string `json:"sku"`
	Quantity  int    `json:"quantity"`
	Reference string `json:"reference"`
}

// StockReserveResponse representa la respuesta de reserva de stock
type StockReserveResponse struct {
	SKU          string `json:"sku"`
	ReservedQty  int    `json:"reserved_qty"`
	RemainingQty int    `json:"remaining_qty"`
	Reference    string `json:"reference"`
}

// StockReleaseRequest representa el request para liberar stock
type StockReleaseRequest struct {
	SKU       string `json:"sku"`
	Quantity  int    `json:"quantity"`
	Reference string `json:"reference"`
}

// StockReleaseResponse representa la respuesta de liberación de stock
type StockReleaseResponse struct {
	SKU          string `json:"sku"`
	ReleasedQty  int    `json:"released_qty"`
	AvailableQty int    `json:"available_qty"`
	ReservedQty  int    `json:"reserved_qty"`
	Reference    string `json:"reference"`
}

// StockConsumeRequest representa el request para consumir stock
type StockConsumeRequest struct {
	SKU       string `json:"sku"`
	Quantity  int    `json:"quantity"`
	Reference string `json:"reference"`
}

// StockConsumeResponse representa la respuesta de consumo de stock
type StockConsumeResponse struct {
	SKU         string `json:"sku"`
	ConsumedQty int    `json:"consumed_qty"`
	ReservedQty int    `json:"reserved_qty"`
	Reference   string `json:"reference"`
}

// StockRevertConsumeRequest representa el request para revertir consumo
type StockRevertConsumeRequest struct {
	SKU       string `json:"sku"`
	Quantity  int    `json:"quantity"`
	Reference string `json:"reference"`
}

// StockRevertConsumeResponse representa la respuesta de reversión
type StockRevertConsumeResponse struct {
	SKU          string `json:"sku"`
	RevertedQty  int    `json:"reverted_qty"`
	AvailableQty int    `json:"available_qty"`
	Reference    string `json:"reference"`
}

// StockClient cliente HTTP para comunicarse con stock-service vía Kong
type StockClient struct {
	httpClient *http.Client
	kongURL    string
	stockPath  string
}

// NewStockClient crea una nueva instancia del cliente
func NewStockClient() *StockClient {
	kongURL := os.Getenv("KONG_INTERNAL_URL")
	if kongURL == "" {
		kongURL = "http://kong:8000" // Default para entorno Docker
	}

	stockPath := os.Getenv("STOCK_SERVICE_PATH")
	if stockPath == "" {
		stockPath = "/stock" // Default
	}

	return &StockClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		kongURL:   kongURL,
		stockPath: stockPath,
	}
}

// ValidateStock valida disponibilidad de stock vía Kong usando GET /availability
func (c *StockClient) ValidateStock(tenantID, authToken, sku string, quantity int) (*StockAvailabilityResponse, bool, error) {
	// Construir URL completa vía Kong con query parameter
	url := fmt.Sprintf("%s%s/api/v1/availability?sku=%s", c.kongURL, c.stockPath, sku)

	// Crear request HTTP
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, false, fmt.Errorf("error creating request: %w", err)
	}

	// Headers obligatorios
	req.Header.Set("X-Tenant-ID", tenantID)
	
	// Pasar Authorization si existe
	if authToken != "" {
		req.Header.Set("Authorization", authToken)
	}

	// Ejecutar request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("error calling stock-service: %w", err)
	}
	defer resp.Body.Close()

	// Leer response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, false, fmt.Errorf("error reading response: %w", err)
	}

	// Verificar status code
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("stock-service returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var stockResp StockAvailabilityResponse
	if err := json.Unmarshal(body, &stockResp); err != nil {
		return nil, false, fmt.Errorf("error unmarshalling response: %w", err)
	}

	// Determinar si hay suficiente stock
	hasEnoughStock := stockResp.AvailableQuantity >= float64(quantity)

	return &stockResp, hasEnoughStock, nil
}

// ReserveStock reserva stock vía Kong usando POST /reserve
func (c *StockClient) ReserveStock(tenantID, authToken, sku string, quantity int, reference string) (*StockReserveResponse, error) {
	// Preparar request body
	reqBody := StockReserveRequest{
		SKU:       sku,
		Quantity:  quantity,
		Reference: reference,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request: %w", err)
	}

	// Construir URL completa vía Kong
	url := fmt.Sprintf("%s%s/api/v1/reserve", c.kongURL, c.stockPath)

	// Crear request HTTP
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Headers obligatorios
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID)

	// Pasar Authorization si existe
	if authToken != "" {
		req.Header.Set("Authorization", authToken)
	}

	// Ejecutar request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling stock-service: %w", err)
	}
	defer resp.Body.Close()

	// Leer response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	// Verificar status code
	if resp.StatusCode == http.StatusConflict {
		return nil, fmt.Errorf("insufficient stock: %s", string(body))
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("stock-service returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var stockResp StockReserveResponse
	if err := json.Unmarshal(body, &stockResp); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %w", err)
	}

	return &stockResp, nil
}

// ReleaseStock libera stock reservado vía Kong usando POST /release
func (c *StockClient) ReleaseStock(tenantID, authToken, sku string, quantity int, reference string) (*StockReleaseResponse, error) {
	// Preparar request body
	reqBody := StockReleaseRequest{
		SKU:       sku,
		Quantity:  quantity,
		Reference: reference,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request: %w", err)
	}

	// Construir URL completa vía Kong
	url := fmt.Sprintf("%s%s/api/v1/release", c.kongURL, c.stockPath)

	// Crear request HTTP
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Headers obligatorios
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID)

	// Pasar Authorization si existe
	if authToken != "" {
		req.Header.Set("Authorization", authToken)
	}

	// Ejecutar request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling stock-service: %w", err)
	}
	defer resp.Body.Close()

	// Leer response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	// Verificar status code
	if resp.StatusCode == http.StatusConflict {
		return nil, fmt.Errorf("insufficient reserved stock: %s", string(body))
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("stock-service returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var stockResp StockReleaseResponse
	if err := json.Unmarshal(body, &stockResp); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %w", err)
	}

	return &stockResp, nil
}

// ConsumeStock consume stock reservado vía Kong usando POST /consume
func (c *StockClient) ConsumeStock(tenantID, authToken, sku string, quantity int, reference string) (*StockConsumeResponse, error) {
	// Preparar request body
	reqBody := StockConsumeRequest{
		SKU:       sku,
		Quantity:  quantity,
		Reference: reference,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request: %w", err)
	}

	// Construir URL completa vía Kong
	url := fmt.Sprintf("%s%s/api/v1/consume", c.kongURL, c.stockPath)

	// Crear request HTTP
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Headers obligatorios
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID)

	// Pasar Authorization si existe
	if authToken != "" {
		req.Header.Set("Authorization", authToken)
	}

	// Ejecutar request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling stock-service: %w", err)
	}
	defer resp.Body.Close()

	// Leer response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	// Verificar status code
	if resp.StatusCode == http.StatusConflict {
		return nil, fmt.Errorf("insufficient reserved stock: %s", string(body))
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("stock-service returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var stockResp StockConsumeResponse
	if err := json.Unmarshal(body, &stockResp); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %w", err)
	}

	return &stockResp, nil
}

// RevertConsume revierte un consumo de stock vía Kong usando POST /revert-consume
func (c *StockClient) RevertConsume(tenantID, authToken, sku string, quantity int, reference string) (*StockRevertConsumeResponse, error) {
	// Preparar request body
	reqBody := StockRevertConsumeRequest{
		SKU:       sku,
		Quantity:  quantity,
		Reference: reference,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request: %w", err)
	}

	// Construir URL completa vía Kong
	url := fmt.Sprintf("%s%s/api/v1/revert-consume", c.kongURL, c.stockPath)

	// Crear request HTTP
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Headers obligatorios
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID)

	// Pasar Authorization si existe
	if authToken != "" {
		req.Header.Set("Authorization", authToken)
	}

	// Ejecutar request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling stock-service: %w", err)
	}
	defer resp.Body.Close()

	// Leer response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	// Verificar status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("stock-service returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var stockResp StockRevertConsumeResponse
	if err := json.Unmarshal(body, &stockResp); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %w", err)
	}

	return &stockResp, nil
}
