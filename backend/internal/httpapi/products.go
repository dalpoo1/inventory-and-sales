package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/stockflow/backend/internal/store"
)

func parseID(r *http.Request) (int, bool) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func parseIntQuery(r *http.Request, key string, fallback int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func clampPageLimit(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}

// listProducts handles GET /api/products?search=&status=&page=&limit=
func listProducts(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		search := r.URL.Query().Get("search")
		status := r.URL.Query().Get("status")
		page, limit := clampPageLimit(parseIntQuery(r, "page", 1), parseIntQuery(r, "limit", 20))

		products, total, err := st.ListProducts(r.Context(), search, status, page, limit)
		if err != nil {
			handleStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"data":  products,
			"page":  page,
			"limit": limit,
			"total": total,
		})
	}
}

// getProduct handles GET /api/products/{id}
func getProduct(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseID(r)
		if !ok {
			writeError(w, http.StatusBadRequest, "VALIDATION", "invalid product id", nil, "", nil)
			return
		}
		p, err := st.GetProduct(r.Context(), id)
		if err != nil {
			handleStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, p)
	}
}

type createProductReq struct {
	SKU   string `json:"sku"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

// createProduct handles POST /api/products
func createProduct(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createProductReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION", "invalid JSON body", nil, "", nil)
			return
		}

		fields := map[string]string{}
		if strings.TrimSpace(req.SKU) == "" {
			fields["sku"] = "sku is required"
		}
		if strings.TrimSpace(req.Name) == "" {
			fields["name"] = "name is required"
		}
		if req.Price < 0 {
			fields["price"] = "price must be >= 0"
		}
		if len(fields) > 0 {
			writeError(w, http.StatusBadRequest, "VALIDATION", "validation failed", fields, "", nil)
			return
		}

		p, err := st.CreateProduct(r.Context(), strings.TrimSpace(req.SKU), strings.TrimSpace(req.Name), req.Price)
		if err != nil {
			handleStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, p)
	}
}

type updateProductReq struct {
	SKU      *string `json:"sku"`
	Name     *string `json:"name"`
	Price    *int    `json:"price"`
	IsActive *bool   `json:"isActive"`
}

// updateProduct handles PUT /api/products/{id}
func updateProduct(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseID(r)
		if !ok {
			writeError(w, http.StatusBadRequest, "VALIDATION", "invalid product id", nil, "", nil)
			return
		}

		var req updateProductReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION", "invalid JSON body", nil, "", nil)
			return
		}

		fields := map[string]string{}
		if req.SKU != nil && strings.TrimSpace(*req.SKU) == "" {
			fields["sku"] = "sku must not be empty"
		}
		if req.Name != nil && strings.TrimSpace(*req.Name) == "" {
			fields["name"] = "name must not be empty"
		}
		if req.Price != nil && *req.Price < 0 {
			fields["price"] = "price must be >= 0"
		}
		if len(fields) > 0 {
			writeError(w, http.StatusBadRequest, "VALIDATION", "validation failed", fields, "", nil)
			return
		}

		input := store.UpdateProductInput{SKU: req.SKU, Name: req.Name, Price: req.Price, IsActive: req.IsActive}
		p, err := st.UpdateProduct(r.Context(), id, input)
		if err != nil {
			handleStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, p)
	}
}

// deleteProduct handles DELETE /api/products/{id}
func deleteProduct(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseID(r)
		if !ok {
			writeError(w, http.StatusBadRequest, "VALIDATION", "invalid product id", nil, "", nil)
			return
		}
		if err := st.DeleteProduct(r.Context(), id); err != nil {
			handleStoreError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

type adjustStockReq struct {
	Type string  `json:"type"`
	Qty  int     `json:"qty"`
	Note *string `json:"note"`
}

// adjustStock handles POST /api/products/{id}/stock
func adjustStock(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseID(r)
		if !ok {
			writeError(w, http.StatusBadRequest, "VALIDATION", "invalid product id", nil, "", nil)
			return
		}

		var req adjustStockReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION", "invalid JSON body", nil, "", nil)
			return
		}

		fields := map[string]string{}
		if req.Type != "in" && req.Type != "out" {
			fields["type"] = "type must be 'in' or 'out'"
		}
		if req.Qty <= 0 {
			fields["qty"] = "qty must be a positive integer"
		}
		if len(fields) > 0 {
			writeError(w, http.StatusBadRequest, "VALIDATION", "validation failed", fields, "", nil)
			return
		}

		p, err := st.AdjustStock(r.Context(), id, req.Type, req.Qty, req.Note)
		if err != nil {
			handleStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, p)
	}
}

// listMovements handles GET /api/products/{id}/movements
func listMovements(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseID(r)
		if !ok {
			writeError(w, http.StatusBadRequest, "VALIDATION", "invalid product id", nil, "", nil)
			return
		}
		movements, err := st.ListMovements(r.Context(), id)
		if err != nil {
			handleStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": movements})
	}
}

// getSummary handles GET /api/summary
func getSummary(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sum, err := st.GetSummary(r.Context())
		if err != nil {
			handleStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, sum)
	}
}
