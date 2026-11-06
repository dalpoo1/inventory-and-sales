package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stockflow/backend/internal/model"
)

// OrderItemInput is a requested order line. Unit price is never supplied by
// the client; it is read from the products table.
type OrderItemInput struct {
	SKU string
	Qty int
}

type productRef struct {
	SKU   string
	ID    int
	Price int
}

// CreateOrder validates every line, then in a single transaction inserts the
// order and its items, atomically decrements stock for each SKU (guarded by
// "stock >= qty" inside the UPDATE), and writes an "out" movement per line.
// Any failure rolls the whole order back: either it exists fully or not at all.
func (s *Store) CreateOrder(ctx context.Context, customer string, items []OrderItemInput, idempotencyKey *string) (*model.Order, error) {
	if idempotencyKey != nil && *idempotencyKey != "" {
		if existing, err := s.orderByIdempotencyKey(ctx, *idempotencyKey); err == nil && existing != nil {
			return existing, nil
		}
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Resolve SKUs to product ids + prices (validation, not a stock decision).
	skus := make([]string, len(items))
	for i, it := range items {
		skus[i] = it.SKU
	}
	rows, err := tx.Query(ctx, `SELECT sku, id, price FROM products WHERE sku = ANY($1)`, skus)
	if err != nil {
		return nil, err
	}
	bySKU := make(map[string]productRef, len(items))
	for rows.Next() {
		var ref productRef
		if err := rows.Scan(&ref.SKU, &ref.ID, &ref.Price); err != nil {
			rows.Close()
			return nil, err
		}
		bySKU[ref.SKU] = ref
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var unknown []string
	seenUnknown := map[string]bool{}
	for _, sku := range skus {
		if _, ok := bySKU[sku]; !ok && !seenUnknown[sku] {
			unknown = append(unknown, sku)
			seenUnknown[sku] = true
		}
	}
	if len(unknown) > 0 {
		return nil, &UnknownSKUError{SKUs: unknown}
	}

	total := 0
	for _, it := range items {
		total += bySKU[it.SKU].Price * it.Qty
	}

	var idemKey any
	if idempotencyKey != nil && *idempotencyKey != "" {
		idemKey = *idempotencyKey
	}

	var orderID int
	var createdAt time.Time
	err = tx.QueryRow(ctx,
		`INSERT INTO sales_orders (customer, status, total, idempotency_key)
		 VALUES ($1, 'CREATED', $2, $3) RETURNING id, created_at`,
		customer, total, idemKey).Scan(&orderID, &createdAt)
	if err != nil {
		if isUniqueViolation(err) && idempotencyKey != nil && *idempotencyKey != "" {
			_ = tx.Rollback(ctx)
			if existing, ferr := s.orderByIdempotencyKey(ctx, *idempotencyKey); ferr == nil && existing != nil {
				return existing, nil
			}
		}
		return nil, err
	}

	for _, it := range items {
		if _, err := tx.Exec(ctx,
			`INSERT INTO sales_order_items (order_id, sku, qty, unit_price) VALUES ($1, $2, $3, $4)`,
			orderID, it.SKU, it.Qty, bySKU[it.SKU].Price); err != nil {
			return nil, err
		}
	}

	for _, it := range items {
		var returnedSKU string
		err := tx.QueryRow(ctx,
			`UPDATE products SET stock = stock - $1 WHERE sku = $2 AND stock >= $1 RETURNING sku`,
			it.Qty, it.SKU).Scan(&returnedSKU)
		if errors.Is(err, pgx.ErrNoRows) {
			var exists bool
			if serr := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM products WHERE sku = $1)`, it.SKU).Scan(&exists); serr != nil {
				return nil, serr
			}
			if !exists {
				return nil, &UnknownSKUError{SKUs: []string{it.SKU}}
			}
			var available int
			if serr := tx.QueryRow(ctx, `SELECT stock FROM products WHERE sku = $1`, it.SKU).Scan(&available); serr != nil {
				return nil, serr
			}
			return nil, &InsufficientStockError{SKU: it.SKU, Available: available}
		}
		if err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO stock_movements (product_id, type, qty, note) VALUES ($1, 'out', $2, $3)`,
			bySKU[it.SKU].ID, it.Qty, fmt.Sprintf("Order #%d", orderID)); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.GetOrder(ctx, orderID)
}

// GetOrder returns an order with its items.
func (s *Store) GetOrder(ctx context.Context, id int) (*model.Order, error) {
	var o model.Order
	err := s.pool.QueryRow(ctx,
		`SELECT id, customer, status, total, created_at, cancelled_at FROM sales_orders WHERE id = $1`, id).
		Scan(&o.ID, &o.Customer, &o.Status, &o.Total, &o.CreatedAt, &o.CancelledAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	rows, err := s.pool.Query(ctx,
		`SELECT id, sku, qty, unit_price FROM sales_order_items WHERE order_id = $1 ORDER BY id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	o.Items = make([]model.OrderItem, 0)
	for rows.Next() {
		var it model.OrderItem
		if err := rows.Scan(&it.ID, &it.SKU, &it.Qty, &it.UnitPrice); err != nil {
			return nil, err
		}
		o.Items = append(o.Items, it)
	}
	return &o, rows.Err()
}

// ListOrders returns a page of orders (newest first) with their items.
func (s *Store) ListOrders(ctx context.Context, page, limit int) ([]model.Order, int, error) {
	var total int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM sales_orders`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := s.pool.Query(ctx,
		`SELECT id, customer, status, total, created_at, cancelled_at
		 FROM sales_orders ORDER BY created_at DESC, id DESC LIMIT $1 OFFSET $2`,
		limit, (page-1)*limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	orders := make([]model.Order, 0, limit)
	index := map[int]int{}
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.ID, &o.Customer, &o.Status, &o.Total, &o.CreatedAt, &o.CancelledAt); err != nil {
			return nil, 0, err
		}
		o.Items = make([]model.OrderItem, 0)
		index[o.ID] = len(orders)
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if len(orders) > 0 {
		ids := make([]int, len(orders))
		for i, o := range orders {
			ids[i] = o.ID
		}
		itemRows, err := s.pool.Query(ctx,
			`SELECT order_id, id, sku, qty, unit_price FROM sales_order_items WHERE order_id = ANY($1) ORDER BY id`, ids)
		if err != nil {
			return nil, 0, err
		}
		for itemRows.Next() {
			var orderID int
			var it model.OrderItem
			if err := itemRows.Scan(&orderID, &it.ID, &it.SKU, &it.Qty, &it.UnitPrice); err != nil {
				itemRows.Close()
				return nil, 0, err
			}
			if i, ok := index[orderID]; ok {
				orders[i].Items = append(orders[i].Items, it)
			}
		}
		itemRows.Close()
		if err := itemRows.Err(); err != nil {
			return nil, 0, err
		}
	}
	return orders, total, nil
}

// CancelOrder marks an order CANCELLED and restores stock in one transaction.
// A FOR UPDATE lock serialises concurrent cancels.
func (s *Store) CancelOrder(ctx context.Context, id int) (*model.Order, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var status string
	err = tx.QueryRow(ctx, `SELECT status FROM sales_orders WHERE id = $1 FOR UPDATE`, id).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if status == "CANCELLED" {
		return nil, ErrOrderAlreadyCancelled
	}

	var cancelledAt time.Time
	if err := tx.QueryRow(ctx,
		`UPDATE sales_orders SET status = 'CANCELLED', cancelled_at = now() WHERE id = $1 RETURNING cancelled_at`, id).
		Scan(&cancelledAt); err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, `SELECT sku, qty FROM sales_order_items WHERE order_id = $1`, id)
	if err != nil {
		return nil, err
	}
	type line struct {
		sku string
		qty int
	}
	var lines []line
	for rows.Next() {
		var l line
		if err := rows.Scan(&l.sku, &l.qty); err != nil {
			rows.Close()
			return nil, err
		}
		lines = append(lines, l)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, l := range lines {
		if _, err := tx.Exec(ctx, `UPDATE products SET stock = stock + $1 WHERE sku = $2`, l.qty, l.sku); err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO stock_movements (product_id, type, qty, note)
			 SELECT id, 'in', $1, $2 FROM products WHERE sku = $3`,
			l.qty, fmt.Sprintf("Cancel order #%d", id), l.sku); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.GetOrder(ctx, id)
}

func (s *Store) orderByIdempotencyKey(ctx context.Context, key string) (*model.Order, error) {
	var id int
	err := s.pool.QueryRow(ctx, `SELECT id FROM sales_orders WHERE idempotency_key = $1`, key).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return s.GetOrder(ctx, id)
}
