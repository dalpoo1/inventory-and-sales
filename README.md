# StockFlow

A warehouse supervisor's single app to **manage products, adjust stock, and record sales orders that consume stock**. A full vertical slice: **PostgreSQL → Go API → React UI**.

> This is a software-engineering exercise: the API contract, cross-stack error handling, and data safety under concurrency matter more than visual polish.

---

## Stack

| Layer     | Tech                                                              |
|-----------|-------------------------------------------------------------------|
| Database  | PostgreSQL 16 (migrations + seed applied automatically)           |
| Backend   | Go 1.22+ (`net/http` + `chi`, `pgx/v5`)                           |
| Frontend  | React 18 + TypeScript (strict) + Vite + Tailwind + TanStack Query |
| Tests     | Go (`testing` + testcontainers), Vitest                           |

## Repository layout

```
stockflow/
├── backend/
│   ├── cmd/server/          # entrypoint (--migrate flag for local dev)
│   ├── internal/
│   │   ├── config/          # env config
│   │   ├── model/           # domain types (JSON contract)
│   │   ├── store/           # data access + transactional business rules
│   │   ├── httpapi/         # chi router, handlers, error envelope
│   │   └── migrate/         # in-process migration runner
│   ├── migrations/          # 0001_schema.sql, 0002_seed.sql (also mounted into Postgres)
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── api/             # types.ts, client.ts (error parsing), index.ts (all endpoints)
│   │   ├── components/      # table, modals, badges, cards
│   │   ├── pages/           # Products, Orders, OrderDetail
│   │   ├── hooks/           # useDebounce
│   │   └── utils/           # format.ts, validation.ts (+ unit tests)
│   └── Dockerfile + nginx.conf
├── docker-compose.yml       # db + backend + frontend, one command
└── README.md
```

---

## Prerequisites

- **Docker** (Desktop, with the Compose plugin) — for the one-command run and the backend integration tests.
- **Go 1.22 or newer** — for local backend development and tests.
- **Node 18+** — for local frontend development.

> The backend `go.mod` targets **Go 1.22**, so any modern toolchain (1.22 through 1.27+) builds it. Dependencies are pinned to versions compatible with 1.22.

## Quick start (≤ 5 minutes)

```bash
docker compose up -d --build
```

This builds and runs **Postgres (migrated + seeded)**, the **Go API**, and the **React app (nginx)**:

| Service   | URL                              |
|-----------|----------------------------------|
| Frontend  | http://localhost:5173            |
| Backend   | http://localhost:8080            |
| Postgres  | localhost:5432 (user/pass/db: `stockflow`) |

The frontend (nginx) proxies `/api` to the backend, and the backend connects to `db`. No other setup needed.

**Teardown:** `docker compose down` (add `-v` to also wipe the database volume and re-apply migrations/seed on next `up`).

---

## Local development (run backend + frontend on your machine)

```bash
# 1. Database (migrations + seed apply automatically on first boot)
docker compose up -d db

# 2. Backend (or: go run ./cmd/server -migrate  to apply migrations without the container)
cd backend
go run ./cmd/server

# 3. Frontend (Vite proxies /api → localhost:8080, so no CORS)
cd frontend
npm install
npm run dev
```

> **Windows note:** running `go test -race` requires a C compiler (cgo). Install
> MinGW-w64/TDM-GCC, or run the race test inside WSL2/Linux:
> `CGO_ENABLED=1 go test -race ./internal/store/`.

---

## API

All responses are JSON. Errors use one envelope:

```json
{ "error": { "code": "INSUFFICIENT_STOCK", "sku": "SKU-0002", "available": 6 } }
```

### Endpoints

| Method | Path                          | Description                                  |
|--------|-------------------------------|----------------------------------------------|
| GET    | `/api/products`               | List, `?search=&status=all|active|inactive|low&page=&limit=` |
| POST   | `/api/products`               | Create (sku unique, name required, price ≥ 0) |
| PUT    | `/api/products/{id}`          | Partial update (sku/name/price/isActive); 404 if missing |
| DELETE | `/api/products/{id}`          | 204; **409 `PRODUCT_HAS_HISTORY`** if it has movement history |
| POST   | `/api/products/{id}/stock`    | `{"type":"in"|"out","qty":n,"note":"…"}`; 409 `INSUFFICIENT_STOCK` on overdraw |
| GET    | `/api/products/{id}/movements`| History, newest first (`{"data":[…]}`)        |
| GET    | `/api/summary`                | Dashboard summary (extension)                |
| GET    | `/api/orders`                 | List orders, `?page=&limit=` (extension)     |
| POST   | `/api/orders`                 | Create (see rules below); 201                |
| GET    | `/api/orders/{id}`            | Order + items + status + total               |
| POST   | `/api/orders/{id}/cancel`     | Restore stock; 409 `ORDER_ALREADY_CANCELLED` |

### Error codes

| HTTP | code                     | context                                        |
|------|--------------------------|------------------------------------------------|
| 400  | `VALIDATION`             | bad input (with `message` and/or `fields`)     |
| 404  | `NOT_FOUND`              | missing product/order                          |
| 409  | `INSUFFICIENT_STOCK`     | overdraw (`sku`, `available`)                  |
| 409  | `PRODUCT_HAS_HISTORY`    | delete on a product with movement history      |
| 409  | `ORDER_ALREADY_CANCELLED`| cancel on an already-cancelled order           |
| 500  | `INTERNAL`               | unexpected server error                        |

### curl examples

```bash
# List products (search + filter + page)
curl 'http://localhost:8080/api/products?search=beras&status=active&page=1&limit=20'

# Create a product
curl -X POST http://localhost:8080/api/products \
  -H 'Content-Type: application/json' \
  -d '{"sku":"SKU-0100","name":"New Item","price":50000}'

# Adjust stock (out, insufficient → 409)
curl -X POST http://localhost:8080/api/products/2/stock \
  -H 'Content-Type: application/json' \
  -d '{"type":"out","qty":999,"note":"damaged"}'

# Movement history
curl http://localhost:8080/api/products/1/movements

# Create an order (server computes the total and unit prices)
curl -X POST http://localhost:8080/api/orders \
  -H 'Content-Type: application/json' \
  -d '{"customer":"Toko Maju","items":[{"sku":"SKU-0001","qty":2},{"sku":"SKU-0006","qty":1}]}'

# Get an order
curl http://localhost:8080/api/orders/1

# Cancel an order (restores stock)
curl -X POST http://localhost:8080/api/orders/1/cancel

# Delete a product (409 if it has history)
curl -X DELETE http://localhost:8080/api/products/12

# Idempotent order creation (retries never double-order)
curl -X POST http://localhost:8080/api/orders \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: 8b4f9c2e-...' \
  -d '{"customer":"Toko Maju","items":[{"sku":"SKU-0001","qty":1}]}'
```

---

## Order creation rules (the core)

1. **Validate before writing:** unknown sku, qty ≤ 0, duplicate skus, empty items → 400 with clear messages.
2. **`unit_price`** always comes from the `products` table; the client never supplies prices. `total` is computed server-side.
3. **Atomic stock decrement** for all items in a single transaction: the order exists fully or not at all.
4. Insufficient stock → 409 `INSUFFICIENT_STOCK` with `sku` + `available`, and **no stock has changed**.
5. **Concurrency-safe:** the guard lives inside the `UPDATE` (`… SET stock = stock - $qty WHERE sku = $sku AND stock >= $qty`), never SELECT-then-UPDATE.

---

## Design decisions

- **Money is `INTEGER` (IDR).** No floating point; the UI formats with `Rp 1.234.567`.
- **`last_movement_at` is derived**, not stored: `MAX(created_at)` from `stock_movements` in the products query. Deriving avoids a denormalised column that can drift; the movement table is the single source of truth.
- **`stock_movements` is an immutable ledger.** History is never deleted (DELETE on a product is refused with `PRODUCT_HAS_HISTORY`). Every stock change — manual adjustment *and* every order line and cancellation — writes a movement, so the ledger always reconciles with `products.stock`.
- **Order creation writes `out` movements** (note `"Order #n"`) and **cancellation writes `in` movements** (note `"Cancel order #n"`), keeping a complete audit trail.
- **Row-locking vs atomic UPDATE.** Stock decrements use a single guarded `UPDATE` so concurrent orders for the same SKU can never drive stock negative. Cancellation uses `SELECT … FOR UPDATE` on the order row to serialise concurrent cancels.
- **`sales_order_items.unit_price`** is a price snapshot at order time, so historical orders keep their original pricing even if a product's price later changes.
- **Error envelope** is uniform and machine-readable; the React client parses it once in `api/client.ts` and surfaces 409s (with `available`) directly in the UI.
- **Idempotency-Key** (bonus): `POST /api/orders` accepts an `Idempotency-Key` header backed by a `UNIQUE` column; a retried request returns the original order instead of double-ordering. The frontend generates and reuses a key per submit attempt.

---

## Tests

### Backend

```bash
cd backend
go test ./...            # unit + integration (testcontainers; needs Docker)
go test -race ./internal/store/   # concurrency test (needs cgo on Windows)
```

Integration tests (real Postgres via testcontainers) cover:

- **Insufficient stock rejected with nothing written** (stock, orders, items, movements all unchanged).
- **Happy path** writes an `out` movement and updates stock, with server-computed total/unit prices.
- **Concurrency** — 10 goroutines order 1 unit against stock of 5 → exactly 5 succeed, 5 get `INSUFFICIENT_STOCK`, final stock `0`, never negative.
- Cancel restores stock; double-cancel → 409; delete with/without history; unknown SKU; adjust-stock overdraw.

### Frontend

```bash
cd frontend
npm test                 # Vitest unit tests for validation + formatting
```

---

## What I would improve (with more time)

- **Rate limiting / auth** — the API is unauthenticated and unthrottled (fine for a test, not production).
- **`GET /api/reports/daily-sales?date=`** and a daily-sales summary card (left as an extension).
- **Pagination on movements** and a UI to browse full movement history.
- **Structured logging + tracing** (currently `chi` request logging only).
- **Migration tooling** — the in-process runner works, but `golang-migrate`/`goose` would add down-migrations and versioned rollback.
- **CI pipeline** running `go test -race` and `vitest` on a Linux runner (avoids the Windows cgo caveat).
- **`npm audit` cleanup** in the frontend dev dependencies.
- Skeleton loading rows and richer styling (kept minimal on purpose).

---

## Honest hours spent

~3 hours of focused, single-session work end-to-end (schema → Go API → tests → React UI → Docker → docs). The concurrency test, error-envelope mapping, and one-command Docker run took the bulk of the time, which is where the evaluation weight sits.
