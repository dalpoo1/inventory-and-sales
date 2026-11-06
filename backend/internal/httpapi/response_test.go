package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stockflow/backend/internal/store"
)

func decodeError(t *testing.T, body string) errorDetail {
	t.Helper()
	var env errorEnvelope
	if err := json.Unmarshal([]byte(body), &env); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	return env.Error
}

func TestHandleStoreError_InsufficientStock(t *testing.T) {
	rr := httptest.NewRecorder()
	handleStoreError(rr, &store.InsufficientStockError{SKU: "SKU-1", Available: 4})

	if rr.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", rr.Code)
	}
	d := decodeError(t, rr.Body.String())
	if d.Code != "INSUFFICIENT_STOCK" || d.SKU != "SKU-1" || d.Available == nil || *d.Available != 4 {
		t.Fatalf("error detail = %+v", d)
	}
}

func TestHandleStoreError_ProductHasHistory(t *testing.T) {
	rr := httptest.NewRecorder()
	handleStoreError(rr, store.ErrProductHasHistory)

	if rr.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", rr.Code)
	}
	d := decodeError(t, rr.Body.String())
	if d.Code != "PRODUCT_HAS_HISTORY" {
		t.Fatalf("code = %s, want PRODUCT_HAS_HISTORY", d.Code)
	}
}

func TestHandleStoreError_NotFound(t *testing.T) {
	rr := httptest.NewRecorder()
	handleStoreError(rr, store.ErrNotFound)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rr.Code)
	}
	d := decodeError(t, rr.Body.String())
	if d.Code != "NOT_FOUND" {
		t.Fatalf("code = %s, want NOT_FOUND", d.Code)
	}
}

func TestHandleStoreError_AlreadyCancelled(t *testing.T) {
	rr := httptest.NewRecorder()
	handleStoreError(rr, store.ErrOrderAlreadyCancelled)

	if rr.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", rr.Code)
	}
	d := decodeError(t, rr.Body.String())
	if d.Code != "ORDER_ALREADY_CANCELLED" {
		t.Fatalf("code = %s, want ORDER_ALREADY_CANCELLED", d.Code)
	}
}
