package entity

import "errors"

var (
	ErrTenantIDRequired         = errors.New("tenant_id is required")
	ErrSKURequired              = errors.New("sku is required")
	ErrInvalidQuantity          = errors.New("quantity must be greater than 0")
	ErrOrderNotFound            = errors.New("order not found")
	ErrOrderNotInCreatedState   = errors.New("order is not in CREATED state")
	ErrOrderNotInConfirmedState = errors.New("order is not in CONFIRMED state")
	ErrOrderMustHaveItems       = errors.New("order must have at least one item")
)
