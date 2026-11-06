package store

import (
	"errors"
	"fmt"
)

// Sentinel and typed errors returned by Store methods. The HTTP layer maps
// these to the API error envelope.
var (
	ErrNotFound             = errors.New("not found")
	ErrProductHasHistory    = errors.New("product has movement history")
	ErrOrderAlreadyCancelled = errors.New("order already cancelled")
)

// InsufficientStockError is returned when a stock decrement would overdraw.
type InsufficientStockError struct {
	SKU       string
	Available int
}

func (e *InsufficientStockError) Error() string {
	if e.SKU != "" {
		return fmt.Sprintf("insufficient stock for %s: available %d", e.SKU, e.Available)
	}
	return fmt.Sprintf("insufficient stock: available %d", e.Available)
}

// UnknownSKUError is returned when an order references a SKU that doesn't exist.
type UnknownSKUError struct {
	SKUs []string
}

func (e *UnknownSKUError) Error() string { return fmt.Sprintf("unknown sku: %v", e.SKUs) }

// ValidationError carries field-level validation messages (HTTP 400).
type ValidationError struct {
	Message string
	Fields  map[string]string
}

func (e *ValidationError) Error() string { return e.Message }
