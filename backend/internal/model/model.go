package model

import "time"

// Product is a stockable item. Money (price) and stock are integers.
type Product struct {
	ID             int        `json:"id"`
	SKU            string     `json:"sku"`
	Name           string     `json:"name"`
	Price          int        `json:"price"`
	Stock          int        `json:"stock"`
	IsActive       bool       `json:"isActive"`
	CreatedAt      time.Time  `json:"createdAt"`
	LastMovementAt *time.Time `json:"lastMovementAt"`
}

// StockMovement is an immutable ledger entry. History is never deleted.
type StockMovement struct {
	ID        int       `json:"id"`
	ProductID int       `json:"productId"`
	Type      string    `json:"type"`
	Qty       int       `json:"qty"`
	Note      *string   `json:"note"`
	CreatedAt time.Time `json:"createdAt"`
}

// Order is a sales order. Status is CREATED or CANCELLED.
type Order struct {
	ID          int         `json:"id"`
	Customer    string      `json:"customer"`
	Status      string      `json:"status"`
	Total       int         `json:"total"`
	CreatedAt   time.Time   `json:"createdAt"`
	CancelledAt *time.Time  `json:"cancelledAt"`
	Items       []OrderItem `json:"items"`
}

// OrderItem is a line on an order with a server-computed unit price snapshot.
type OrderItem struct {
	ID        int    `json:"id"`
	SKU       string `json:"sku"`
	Qty       int    `json:"qty"`
	UnitPrice int    `json:"unitPrice"`
}

// Summary is the dashboard summary, derived from the products table.
type Summary struct {
	TotalProducts   int `json:"totalProducts"`
	ActiveProducts  int `json:"activeProducts"`
	TotalStockValue int `json:"totalStockValue"`
	LowStockCount   int `json:"lowStockCount"`
}
