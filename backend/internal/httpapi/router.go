package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stockflow/backend/internal/store"
)

// NewRouter wires all API routes to the store.
func NewRouter(st *store.Store) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	r.Route("/api", func(r chi.Router) {
		r.Get("/summary", getSummary(st))

		r.Get("/products", listProducts(st))
		r.Post("/products", createProduct(st))

		r.Route("/products/{id}", func(r chi.Router) {
			r.Get("/", getProduct(st))
			r.Put("/", updateProduct(st))
			r.Delete("/", deleteProduct(st))
			r.Post("/stock", adjustStock(st))
			r.Get("/movements", listMovements(st))
		})

		r.Get("/orders", listOrders(st))
		r.Post("/orders", createOrder(st))

		r.Route("/orders/{id}", func(r chi.Router) {
			r.Get("/", getOrder(st))
			r.Post("/cancel", cancelOrder(st))
		})
	})

	return r
}

// corsMiddleware is a permissive fallback; the primary dev path uses the Vite
// proxy and production uses nginx, so CORS is mostly for direct browser hits.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Idempotency-Key")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
