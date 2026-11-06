package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/stockflow/backend/internal/store"
)

type orderItemReq struct {
	SKU string `json:"sku"`
	Qty int    `json:"qty"`
}

type createOrderReq struct {
	Customer string         `json:"customer"`
	Items    []orderItemReq `json:"items"`
}

// createOrder handles POST /api/orders
func createOrder(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createOrderReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION", "invalid JSON body", nil, "", nil)
			return
		}

		fields := map[string]string{}
		if strings.TrimSpace(req.Customer) == "" {
			fields["customer"] = "customer is required"
		}
		if len(req.Items) == 0 {
			fields["items"] = "items must not be empty"
		}

		seen := map[string]bool{}
		for i, it := range req.Items {
			sku := strings.TrimSpace(it.SKU)
			if sku == "" {
				fields[fmt.Sprintf("items[%d].sku", i)] = "sku is required"
			} else if seen[sku] {
				fields[fmt.Sprintf("items[%d].sku", i)] = "duplicate sku: " + sku
			}
			if it.Qty <= 0 {
				fields[fmt.Sprintf("items[%d].qty", i)] = "qty must be a positive integer"
			}
			if sku != "" {
				seen[sku] = true
			}
		}
		if len(fields) > 0 {
			writeError(w, http.StatusBadRequest, "VALIDATION", "validation failed", fields, "", nil)
			return
		}

		items := make([]store.OrderItemInput, len(req.Items))
		for i, it := range req.Items {
			items[i] = store.OrderItemInput{SKU: strings.TrimSpace(it.SKU), Qty: it.Qty}
		}

		var idempotencyKey *string
		if k := r.Header.Get("Idempotency-Key"); strings.TrimSpace(k) != "" {
			k = strings.TrimSpace(k)
			idempotencyKey = &k
		}

		order, err := st.CreateOrder(r.Context(), strings.TrimSpace(req.Customer), items, idempotencyKey)
		if err != nil {
			handleStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, order)
	}
}

// getOrder handles GET /api/orders/{id}
func getOrder(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseID(r)
		if !ok {
			writeError(w, http.StatusBadRequest, "VALIDATION", "invalid order id", nil, "", nil)
			return
		}
		order, err := st.GetOrder(r.Context(), id)
		if err != nil {
			handleStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, order)
	}
}

// listOrders handles GET /api/orders?page=&limit=
func listOrders(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page, limit := clampPageLimit(parseIntQuery(r, "page", 1), parseIntQuery(r, "limit", 20))
		orders, total, err := st.ListOrders(r.Context(), page, limit)
		if err != nil {
			handleStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"data":  orders,
			"page":  page,
			"limit": limit,
			"total": total,
		})
	}
}

// cancelOrder handles POST /api/orders/{id}/cancel
func cancelOrder(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseID(r)
		if !ok {
			writeError(w, http.StatusBadRequest, "VALIDATION", "invalid order id", nil, "", nil)
			return
		}
		order, err := st.CancelOrder(r.Context(), id)
		if err != nil {
			handleStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, order)
	}
}
