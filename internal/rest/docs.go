package rest

import (
	_ "embed"
	"net/http"
)

//go:embed openapi.json
var OpenAPISpec []byte

// RegisterDocsRoutes serves the OpenAPI 3.2.0 JSON spec and interactive Swagger UI.
func RegisterDocsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(OpenAPISpec)
	})

	mux.HandleFunc("GET /docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(swaggerHTML))
	})
}

const swaggerHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Sample Business Service - API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  <style>
    body { margin: 0; background: #fafafa; }
    .banner {
      background: #0a6ed1; color: #fff; padding: 10px 24px;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
      display: flex; justify-content: space-between; align-items: center; font-size: 14px;
    }
    .banner a { color: #fff; text-decoration: underline; margin-left: 16px; }
  </style>
</head>
<body>
  <div class="banner">
    <div><strong>Sample Business Service</strong> &mdash; REST + OpenAPI 3.2.0 + MCP</div>
    <div>
      <a href="/openapi.json">openapi.json</a>
      <a href="/api/v1/summary">KPI Summary</a>
      <a href="/healthz">Health</a>
    </div>
  </div>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = () => {
      SwaggerUIBundle({
        url: '/openapi.json',
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [SwaggerUIBundle.presets.apis]
      });
    };
  </script>
</body>
</html>`
