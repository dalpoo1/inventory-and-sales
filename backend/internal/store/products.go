package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stockflow/backend/internal/model"
)

const productColumns = `p.id, p.sku, p.name, p.price, p.stock, p.is_active, p.created_at,
	(SELECT MAX(m.created_at) FROM stock_movements m WHERE m.product_id = p.id)`

func scanProduct(row pgx.Row) (model.Product, error) {
	var p model.Product
	var lastAt *time.Time
	if err := row.Scan(&p.ID, &p.SKU, &p.Name, &p.Price, &p.Stock, &p.IsActive, &p.CreatedAt, &lastAt); err != nil {
		return p, err
	}
	p.LastMovementAt = lastAt
	return p, nil
}

// buildProductFilters builds the WHERE clause and args for product listing.
// search matches sku or name case-insensitively; status is all|active|inactive|low.
func buildProductFilters(search, status string) (string, []any) {
	var clauses []string
	var args []any
	arg := func(v any) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}

	if strings.TrimSpace(search) != "" {
		like := "%" + search + "%"
		clauses = append(clauses, "(p.sku ILIKE "+arg(like)+" OR p.name ILIKE "+arg(like)+")")
	}
	switch status {
	case "active":
		clauses = append(clauses, "p.is_active = TRUE")
	case "inactive":
		clauses = append(clauses, "p.is_active = FALSE")
	case "low":
		// Match the summary's definition of "low stock": active products only.
		clauses = append(clauses, "p.is_active = TRUE AND p.stock <= 10")
	}

	where := ""
	if len(clauses) > 0 {
		where = "WHERE " + strings.Join(clauses, " AND ")
	}
	return where, args
}

// ListProducts returns a page of products matching the filters.
func (s *Store) ListProducts(ctx context.Context, search, status string, page, limit int) ([]model.Product, int, error) {
	where, filterArgs := buildProductFilters(search, status)

	var total int
	if err := s.pool.QueryRow(ctx, "SELECT count(*) FROM products p "+where, filterArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := "SELECT " + productColumns + " FROM products p " + where + " ORDER BY p.id LIMIT $" +
		fmt.Sprintf("%d", len(filterArgs)+1) + " OFFSET $" + fmt.Sprintf("%d", len(filterArgs)+2)

	args := append(filterArgs, limit, (page-1)*limit)
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	products := make([]model.Product, 0, limit)
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, 0, err
		}
		products = append(products, p)
	}
	return products, total, rows.Err()
}

// GetProduct returns a single product by id.
func (s *Store) GetProduct(ctx context.Context, id int) (*model.Product, error) {
	row := s.pool.QueryRow(ctx, "SELECT "+productColumns+" FROM products p WHERE p.id = $1", id)
	p, err := scanProduct(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// CreateProduct inserts a product. Uniqueness is enforced by the DB; a unique
// violation is translated to a ValidationError.
func (s *Store) CreateProduct(ctx context.Context, sku, name string, price int) (*model.Product, error) {
	row := s.pool.QueryRow(ctx,
		`INSERT INTO products (sku, name, price) VALUES ($1, $2, $3)
		 RETURNING id, sku, name, price, stock, is_active, created_at`,
		sku, name, price)
	var p model.Product
	if err := row.Scan(&p.ID, &p.SKU, &p.Name, &p.Price, &p.Stock, &p.IsActive, &p.CreatedAt); err != nil {
		if isUniqueViolation(err) {
			return nil, &ValidationError{Message: "sku already exists", Fields: map[string]string{"sku": "already exists"}}
		}
		return nil, err
	}
	return &p, nil
}

// UpdateProductInput is a partial update; nil fields are left unchanged.
type UpdateProductInput struct {
	SKU      *string
	Name     *string
	Price    *int
	IsActive *bool
}

// UpdateProduct applies a partial update and returns the refreshed product.
func (s *Store) UpdateProduct(ctx context.Context, id int, in UpdateProductInput) (*model.Product, error) {
	var sets []string
	var args []any
	arg := func(v any) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}
	if in.SKU != nil {
		sets = append(sets, "sku = "+arg(*in.SKU))
	}
	if in.Name != nil {
		sets = append(sets, "name = "+arg(*in.Name))
	}
	if in.Price != nil {
		sets = append(sets, "price = "+arg(*in.Price))
	}
	if in.IsActive != nil {
		sets = append(sets, "is_active = "+arg(*in.IsActive))
	}
	if len(sets) == 0 {
		return s.GetProduct(ctx, id)
	}

	query := "UPDATE products SET " + strings.Join(sets, ", ") + " WHERE id = " + arg(id)
	tag, err := s.pool.Exec(ctx, query, args...)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, &ValidationError{Message: "sku already exists", Fields: map[string]string{"sku": "already exists"}}
		}
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrNotFound
	}
	return s.GetProduct(ctx, id)
}

// DeleteProduct removes a product only if it has no movement history.
// The existence + history checks run in a transaction with a row lock so a
// concurrent stock adjustment cannot sneak a movement in mid-delete.
func (s *Store) DeleteProduct(ctx context.Context, id int) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM products WHERE id = $1 FOR UPDATE)`, id).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}

	var hasHistory bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM stock_movements WHERE product_id = $1)`, id).Scan(&hasHistory); err != nil {
		return err
	}
	if hasHistory {
		return ErrProductHasHistory
	}

	if _, err := tx.Exec(ctx, `DELETE FROM products WHERE id = $1`, id); err != nil {
		// A product referenced by an order (FK) has order history, hence
		// movement history; guard as a fallback in case of ordering edge cases.
		if isForeignKeyViolation(err) {
			return ErrProductHasHistory
		}
		return err
	}
	return tx.Commit(ctx)
}

// AdjustStock applies an "in" or "out" adjustment and writes a movement row in
// the same transaction. For "out", the UPDATE itself is the guard against
// overdraw: stock is decremented only if stock >= qty.
func (s *Store) AdjustStock(ctx context.Context, id int, typ string, qty int, note *string) (*model.Product, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var query string
	if typ == "in" {
		query = `UPDATE products SET stock = stock + $1 WHERE id = $2
		         RETURNING id, sku, name, price, stock, is_active, created_at`
	} else {
		query = `UPDATE products SET stock = stock - $1 WHERE id = $2 AND stock >= $1
		         RETURNING id, sku, name, price, stock, is_active, created_at`
	}

	var p model.Product
	err = tx.QueryRow(ctx, query, qty, id).Scan(&p.ID, &p.SKU, &p.Name, &p.Price, &p.Stock, &p.IsActive, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		var exists bool
		if serr := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)`, id).Scan(&exists); serr != nil {
			return nil, serr
		}
		if !exists {
			return nil, ErrNotFound
		}
		var available int
		if serr := tx.QueryRow(ctx, `SELECT stock FROM products WHERE id = $1`, id).Scan(&available); serr != nil {
			return nil, serr
		}
		return nil, &InsufficientStockError{Available: available}
	}
	if err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO stock_movements (product_id, type, qty, note) VALUES ($1, $2, $3, $4)`,
		id, typ, qty, note); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	now := time.Now()
	p.LastMovementAt = &now
	return &p, nil
}

// ListMovements returns a product's movement history, newest first.
func (s *Store) ListMovements(ctx context.Context, productID int) ([]model.StockMovement, error) {
	var exists bool
	if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)`, productID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrNotFound
	}

	rows, err := s.pool.Query(ctx,
		`SELECT id, product_id, type, qty, note, created_at
		 FROM stock_movements WHERE product_id = $1 ORDER BY created_at DESC, id DESC`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	movements := make([]model.StockMovement, 0)
	for rows.Next() {
		var m model.StockMovement
		if err := rows.Scan(&m.ID, &m.ProductID, &m.Type, &m.Qty, &m.Note, &m.CreatedAt); err != nil {
			return nil, err
		}
		movements = append(movements, m)
	}
	return movements, rows.Err()
}

// GetSummary computes the dashboard summary from the products table.
// Low stock is defined as active products with stock <= 10.
func (s *Store) GetSummary(ctx context.Context) (*model.Summary, error) {
	var sum model.Summary
	err := s.pool.QueryRow(ctx, `
		SELECT
			count(*) AS total_products,
			count(*) FILTER (WHERE is_active) AS active_products,
			coalesce(sum(price * stock), 0) AS total_stock_value,
			count(*) FILTER (WHERE is_active AND stock <= 10) AS low_stock_count
		FROM products`).Scan(&sum.TotalProducts, &sum.ActiveProducts, &sum.TotalStockValue, &sum.LowStockCount)
	if err != nil {
		return nil, err
	}
	return &sum, nil
}

// isUniqueViolation reports whether err is a PostgreSQL unique violation (23505).
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// isForeignKeyViolation reports whether err is a PostgreSQL FK violation (23503).
func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
