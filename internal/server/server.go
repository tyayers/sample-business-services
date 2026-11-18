package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/tyayers/sample-business-services/internal/domain"
	"github.com/tyayers/sample-business-services/internal/mcp"
	"github.com/tyayers/sample-business-services/internal/rest"
	"github.com/tyayers/sample-business-services/public"
)

// New builds the complete HTTP handler with static frontend, REST, OpenAPI docs, and MCP endpoints.
func New(store *domain.Store) http.Handler {
	mux := http.NewServeMux()

	// Health & API index
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /api/v1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(map[string]any{
			"service": "sample-business-service",
			"version": "1.0.0",
			"endpoints": map[string]string{
				"explorer":       "/",
				"docs":           "/docs",
				"openapi":        "/openapi.json",
				"mcp":            "/mcp",
				"health":         "/healthz",
				"summary":        "/api/v1/summary",
				"customers":      "/api/v1/customers",
				"materials":      "/api/v1/materials",
				"vendors":        "/api/v1/vendors",
				"salesOrders":    "/api/v1/sales-orders",
				"invoices":       "/api/v1/invoices",
				"purchaseOrders": "/api/v1/purchase-orders",
			},
		})
	})

	// REST API + OpenAPI docs
	restHandler := rest.NewHandler(store)
	restHandler.RegisterRoutes(mux)
	rest.RegisterDocsRoutes(mux)

	// MCP Streamable HTTP endpoint
	mcpServer := mcp.NewServer(store)
	mcpServer.RegisterRoutes(mux)

	// Static frontend in public/
	mux.Handle("GET /", http.FileServer(http.FS(public.Assets)))

	// Middleware chain: CORS -> Logging -> Auth (placeholder) -> Router
	return corsMiddleware(loggingMiddleware(authMiddleware(mux)))
}

// authMiddleware is a placeholder hook for future authentication (JWT / OAuth2 / API Key).
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Add authentication verification here when ready.
		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Mcp-Session-Id")
		w.Header().Set("Access-Control-Expose-Headers", "Mcp-Session-Id")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		log.Printf("%s %s %d (%s)", r.Method, r.URL.Path, rec.status, time.Since(start).Round(time.Microsecond))
	})
}
