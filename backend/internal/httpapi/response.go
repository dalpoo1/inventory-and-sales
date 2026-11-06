package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/stockflow/backend/internal/store"
)

// errorDetail is the "error" object of the API error envelope.
type errorDetail struct {
	Code      string            `json:"code"`
	Message   string            `json:"message,omitempty"`
	Fields    map[string]string `json:"fields,omitempty"`
	SKU       string            `json:"sku,omitempty"`
	Available *int              `json:"available,omitempty"`
}

type errorEnvelope struct {
	Error errorDetail `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string, fields map[string]string, sku string, available *int) {
	writeJSON(w, status, errorEnvelope{Error: errorDetail{
		Code:      code,
		Message:   message,
		Fields:    fields,
		SKU:       sku,
		Available: available,
	}})
}

// handleStoreError maps a Store error to the appropriate HTTP response.
func handleStoreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "resource not found", nil, "", nil)
	case errors.Is(err, store.ErrProductHasHistory):
		writeError(w, http.StatusConflict, "PRODUCT_HAS_HISTORY", "", nil, "", nil)
	case errors.Is(err, store.ErrOrderAlreadyCancelled):
		writeError(w, http.StatusConflict, "ORDER_ALREADY_CANCELLED", "order already cancelled", nil, "", nil)
	default:
		var ins *store.InsufficientStockError
		if errors.As(err, &ins) {
			avail := ins.Available
			writeError(w, http.StatusConflict, "INSUFFICIENT_STOCK", "insufficient stock", nil, ins.SKU, &avail)
			return
		}
		var unk *store.UnknownSKUError
		if errors.As(err, &unk) {
			msg := "unknown sku: " + strings.Join(unk.SKUs, ", ")
			writeError(w, http.StatusBadRequest, "VALIDATION", msg, map[string]string{"items": msg}, "", nil)
			return
		}
		var val *store.ValidationError
		if errors.As(err, &val) {
			writeError(w, http.StatusBadRequest, "VALIDATION", val.Message, val.Fields, "", nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL", "internal server error", nil, "", nil)
	}
}
