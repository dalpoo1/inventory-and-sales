package store_test

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stockflow/backend/internal/migrate"
	"github.com/stockflow/backend/internal/model"
	"github.com/stockflow/backend/internal/store"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testStore *store.Store

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Allow running against an already-up database (e.g. `docker compose up db`)
	// via TEST_DATABASE_URL, which is also how the race detector runs on hosts
	// without a local Docker/C toolchain.
	var (
		st      *store.Store
		cleanup = func() {}
	)
	if url := os.Getenv("TEST_DATABASE_URL"); url != "" {
		var err error
		st, err = store.New(ctx, url)
		if err != nil {
			fmt.Fprintf(os.Stderr, "connect to TEST_DATABASE_URL: %v\n", err)
			os.Exit(1)
		}
		cleanup = func() { st.Close() }
	} else {
		var err error
		st, cleanup, err = startPostgres(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skipping integration tests (docker unavailable?): %v\n", err)
			os.Exit(0)
		}
	}
	defer cleanup()

	if err := migrate.Up(ctx, st.Pool()); err != nil {
		fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
		os.Exit(1)
	}
	testStore = st
	os.Exit(m.Run())
}

func startPostgres(ctx context.Context) (*store.Store, func(), error) {
	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("stockflow"),
		postgres.WithUsername("stockflow"),
		postgres.WithPassword("stockflow"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(90*time.Second)),
	)
	if err != nil {
		return nil, nil, err
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, nil, err
	}

	st, err := store.New(ctx, connStr)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, nil, err
	}

	cleanup := func() {
		st.Close()
		_ = container.Terminate(ctx)
	}
	return st, cleanup, nil
}

func resetDB(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	_, err := testStore.Pool().Exec(ctx,
		`TRUNCATE sales_order_items, sales_orders, stock_movements, products RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatalf("reset db: %v", err)
	}
}

// seedProduct inserts a product with a given stock directly (no movement).
func seedProduct(t *testing.T, sku string, price, stock int) model.Product {
	t.Helper()
	ctx := context.Background()
	var p model.Product
	err := testStore.Pool().QueryRow(ctx,
		`INSERT INTO products (sku, name, price, stock) VALUES ($1, $2, $3, $4)
		 RETURNING id, sku, name, price, stock, is_active, created_at`,
		sku, sku+" name", price, stock).
		Scan(&p.ID, &p.SKU, &p.Name, &p.Price, &p.Stock, &p.IsActive, &p.CreatedAt)
	if err != nil {
		t.Fatalf("seed product: %v", err)
	}
	return p
}

func productStock(t *testing.T, sku string) int {
	t.Helper()
	var stock int
	if err := testStore.Pool().QueryRow(context.Background(),
		`SELECT stock FROM products WHERE sku = $1`, sku).Scan(&stock); err != nil {
		t.Fatalf("read stock: %v", err)
	}
	return stock
}

func countRows(t *testing.T, query string, args ...any) int {
	t.Helper()
	var n int
	if err := testStore.Pool().QueryRow(context.Background(), query, args...).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	return n
}

func movementCount(t *testing.T, productID int) int {
	return countRows(t, `SELECT count(*) FROM stock_movements WHERE product_id = $1`, productID)
}

func TestCreateOrder_HappyPath_WritesMovementAndUpdatesStock(t *testing.T) {
	resetDB(t)
	ctx := context.Background()

	p1 := seedProduct(t, "TEST-1", 1000, 10)
	p2 := seedProduct(t, "TEST-2", 500, 5)

	order, err := testStore.CreateOrder(ctx, "Acme", []store.OrderItemInput{
		{SKU: "TEST-1", Qty: 2},
		{SKU: "TEST-2", Qty: 1},
	}, nil)
	if err != nil {
		t.Fatalf("create order: %v", err)
	}

	if order.Total != 2500 {
		t.Fatalf("total = %d, want 2500", order.Total)
	}
	if order.Status != "CREATED" {
		t.Fatalf("status = %s, want CREATED", order.Status)
	}
	if len(order.Items) != 2 {
		t.Fatalf("items = %d, want 2", len(order.Items))
	}
	// unit prices come from the products table, not the client
	if order.Items[0].UnitPrice != 1000 || order.Items[1].UnitPrice != 500 {
		t.Fatalf("unit prices = %d,%d want 1000,500", order.Items[0].UnitPrice, order.Items[1].UnitPrice)
	}

	if got := productStock(t, "TEST-1"); got != 8 {
		t.Fatalf("TEST-1 stock = %d, want 8", got)
	}
	if got := productStock(t, "TEST-2"); got != 4 {
		t.Fatalf("TEST-2 stock = %d, want 4", got)
	}
	if movementCount(t, p1.ID) != 1 {
		t.Fatalf("TEST-1 movements = %d, want 1", movementCount(t, p1.ID))
	}
	if movementCount(t, p2.ID) != 1 {
		t.Fatalf("TEST-2 movements = %d, want 1", movementCount(t, p2.ID))
	}
}

func TestCreateOrder_InsufficientStock_NothingWritten(t *testing.T) {
	resetDB(t)
	ctx := context.Background()

	p := seedProduct(t, "TEST-LOW", 100, 3)

	_, err := testStore.CreateOrder(ctx, "Acme", []store.OrderItemInput{
		{SKU: "TEST-LOW", Qty: 5},
	}, nil)

	ins, ok := err.(*store.InsufficientStockError)
	if !ok {
		t.Fatalf("error = %v, want *InsufficientStockError", err)
	}
	if ins.Available != 3 || ins.SKU != "TEST-LOW" {
		t.Fatalf("insufficient = {sku:%s available:%d}, want {TEST-LOW 3}", ins.SKU, ins.Available)
	}

	// Nothing may have changed: stock, orders, items, movements.
	if got := productStock(t, "TEST-LOW"); got != 3 {
		t.Fatalf("stock = %d, want 3 (unchanged)", got)
	}
	if n := countRows(t, `SELECT count(*) FROM sales_orders`); n != 0 {
		t.Fatalf("orders = %d, want 0", n)
	}
	if n := countRows(t, `SELECT count(*) FROM sales_order_items`); n != 0 {
		t.Fatalf("order items = %d, want 0", n)
	}
	if movementCount(t, p.ID) != 0 {
		t.Fatalf("movements = %d, want 0", movementCount(t, p.ID))
	}
}

func TestCreateOrder_UnknownSKU_Validation(t *testing.T) {
	resetDB(t)
	ctx := context.Background()
	_, err := testStore.CreateOrder(ctx, "Acme", []store.OrderItemInput{
		{SKU: "NOPE", Qty: 1},
	}, nil)
	if _, ok := err.(*store.UnknownSKUError); !ok {
		t.Fatalf("error = %v, want *UnknownSKUError", err)
	}
}

func TestCreateOrder_Concurrent_NeverNegative(t *testing.T) {
	resetDB(t)
	ctx := context.Background()

	p := seedProduct(t, "TEST-RACE", 100, 5)

	const workers = 10
	var wg sync.WaitGroup
	results := make([]error, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := testStore.CreateOrder(ctx, "Acme", []store.OrderItemInput{
				{SKU: "TEST-RACE", Qty: 1},
			}, nil)
			results[i] = err
		}(i)
	}
	wg.Wait()

	successes, insufficient := 0, 0
	for _, err := range results {
		switch {
		case err == nil:
			successes++
		case isInsufficient(err):
			insufficient++
		default:
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if successes != 5 || insufficient != 5 {
		t.Fatalf("successes=%d insufficient=%d, want 5/5", successes, insufficient)
	}
	if got := productStock(t, "TEST-RACE"); got != 0 {
		t.Fatalf("final stock = %d, want 0 (never negative)", got)
	}
	if movementCount(t, p.ID) != 5 {
		t.Fatalf("movements = %d, want 5", movementCount(t, p.ID))
	}
	if n := countRows(t, `SELECT count(*) FROM sales_orders`); n != 5 {
		t.Fatalf("orders = %d, want 5", n)
	}
}

func TestAdjustStock_Out_Insufficient(t *testing.T) {
	resetDB(t)
	ctx := context.Background()
	p := seedProduct(t, "TEST-ADJ", 100, 2)

	_, err := testStore.AdjustStock(ctx, p.ID, "out", 5, nil)
	ins, ok := err.(*store.InsufficientStockError)
	if !ok {
		t.Fatalf("error = %v, want *InsufficientStockError", err)
	}
	if ins.Available != 2 {
		t.Fatalf("available = %d, want 2", ins.Available)
	}
	if got := productStock(t, "TEST-ADJ"); got != 2 {
		t.Fatalf("stock = %d, want 2", got)
	}
	if movementCount(t, p.ID) != 0 {
		t.Fatalf("movements = %d, want 0", movementCount(t, p.ID))
	}
}

func TestAdjustStock_In_WritesMovement(t *testing.T) {
	resetDB(t)
	ctx := context.Background()
	p := seedProduct(t, "TEST-IN", 100, 2)

	updated, err := testStore.AdjustStock(ctx, p.ID, "in", 5, nil)
	if err != nil {
		t.Fatalf("adjust: %v", err)
	}
	if updated.Stock != 7 {
		t.Fatalf("stock = %d, want 7", updated.Stock)
	}
	if movementCount(t, p.ID) != 1 {
		t.Fatalf("movements = %d, want 1", movementCount(t, p.ID))
	}
}

func TestCancelOrder_RestoresStock(t *testing.T) {
	resetDB(t)
	ctx := context.Background()
	p := seedProduct(t, "TEST-CANCEL", 100, 10)

	order, err := testStore.CreateOrder(ctx, "Acme", []store.OrderItemInput{
		{SKU: "TEST-CANCEL", Qty: 3},
	}, nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if got := productStock(t, "TEST-CANCEL"); got != 7 {
		t.Fatalf("stock after order = %d, want 7", got)
	}

	cancelled, err := testStore.CancelOrder(ctx, order.ID)
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if cancelled.Status != "CANCELLED" || cancelled.CancelledAt == nil {
		t.Fatalf("cancelled order = %+v", cancelled)
	}
	if got := productStock(t, "TEST-CANCEL"); got != 10 {
		t.Fatalf("stock after cancel = %d, want 10", got)
	}
	// "in" movement written for the restore
	if movementCount(t, p.ID) != 2 {
		t.Fatalf("movements = %d, want 2 (1 out + 1 in)", movementCount(t, p.ID))
	}
}

func TestCancelOrder_AlreadyCancelled(t *testing.T) {
	resetDB(t)
	ctx := context.Background()
	seedProduct(t, "TEST-2X", 100, 10)

	order, err := testStore.CreateOrder(ctx, "Acme", []store.OrderItemInput{{SKU: "TEST-2X", Qty: 1}}, nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := testStore.CancelOrder(ctx, order.ID); err != nil {
		t.Fatalf("first cancel: %v", err)
	}
	_, err = testStore.CancelOrder(ctx, order.ID)
	if err != store.ErrOrderAlreadyCancelled {
		t.Fatalf("second cancel error = %v, want ErrOrderAlreadyCancelled", err)
	}
}

func TestDeleteProduct_WithHistory(t *testing.T) {
	resetDB(t)
	ctx := context.Background()
	p := seedProduct(t, "TEST-DEL-HIST", 100, 5)
	if _, err := testStore.AdjustStock(ctx, p.ID, "in", 1, nil); err != nil {
		t.Fatalf("adjust: %v", err)
	}
	if err := testStore.DeleteProduct(ctx, p.ID); err != store.ErrProductHasHistory {
		t.Fatalf("delete error = %v, want ErrProductHasHistory", err)
	}
}

func TestDeleteProduct_NoHistory(t *testing.T) {
	resetDB(t)
	ctx := context.Background()
	p := seedProduct(t, "TEST-DEL-NOHIST", 100, 5)
	if err := testStore.DeleteProduct(ctx, p.ID); err != nil {
		t.Fatalf("delete error = %v, want nil", err)
	}
}

func isInsufficient(err error) bool {
	_, ok := err.(*store.InsufficientStockError)
	return ok
}
