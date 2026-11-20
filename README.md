# Sample Business Service

A zero-dependency Go service modeling core enterprise business domains with clean REST APIs, an OpenAPI 3.2.0 specification (plus interactive Swagger UI), and a built-in Model Context Protocol (MCP) endpoint (`/mcp`) — all backed by a shared in-memory enterprise demo dataset.

## Features

- **Zero external dependencies** — pure Go standard library (`net/http`, `encoding/json`, `sync`, `embed`).
- **6 Enterprise Business Domains**:
  - **Customer Master** (`/api/v1/customers`) — credit limits, credit exposure, payment terms, sales orgs, active/blocked status.
  - **Material Master & Inventory** (`/api/v1/materials`) — `FERT` / `HAWA` / `ROH` types, stock levels, reorder points, plants.
  - **Vendor Master** (`/api/v1/vendors`) — suppliers, ratings, payment terms, `ACTIVE` / `ON_HOLD` status.
  - **Sales Orders** (`/api/v1/sales-orders`) — multi-line orders with automatic credit-limit and stock validation/reservation.
  - **Billing & Invoices** (`/api/v1/invoices`) — invoice generation from sales orders and payment recording (releases customer credit).
  - **Procurement / Purchase Orders** (`/api/v1/purchase-orders`) — PO creation and goods receipt (`POST .../receive`), which increments material stock.
- **OpenAPI 3.2.0 + Swagger UI** — `/openapi.json` and `/docs`.
- **MCP Streamable HTTP Endpoint** — `POST /mcp` exposing 21 tools and 2 resources (`business://summary`, `business://openapi.json`) over the same live store.
- **Auth-ready middleware** — placeholder `authMiddleware` hook in `internal/server/server.go` for adding JWT/OAuth2/API-key auth later.

---

## Quick Start

Requires **Go 1.22+** (built and tested with Go 1.26).

```bash
# Run the server (defaults to port 8080)
go run .

# Or specify a custom port
PORT=9090 go run .
```

Once running:
- **Interactive Explorer UI (REST + MCP):** http://localhost:8080/
- **Swagger UI:** http://localhost:8080/docs
- **OpenAPI 3.2.0 Spec:** http://localhost:8080/openapi.json
- **MCP Endpoint:** http://localhost:8080/mcp
- **Health Check:** http://localhost:8080/healthz

Run the test suite:

```bash
go test ./...
```

---

## Deploying to Cloud Run & Apigee

Prerequisites: `go`, `gcloud` CLI (authenticated with permissions for Cloud Run, IAM, IAP, and Apigee), and [`aft`](https://github.com/apigee/apigee-templater).

1. **Copy `env.sample` to `.env` and fill in your values:**

   ```bash
   cp env.sample .env
   ```

   Edit `.env`:
   ```bash
   export GOOGLE_CLOUD_PROJECT=your-gcp-project-id
   export GOOGLE_CLOUD_LOCATION=europe-west1
   export APIGEE_ORG=your-apigee-org          # optional, defaults to GOOGLE_CLOUD_PROJECT
   export APIGEE_ENV=default-dev              # Apigee environment name
   export SERVICE_ACCOUNT=apigee-service      # short name or full SA email
   export IAP_MEMBERS=                        # optional comma-separated principals (defaults to current gcloud user)
   ```

2. **Load `.env` and run `scripts/deploy.sh`:**

   ```bash
   source .env
   ./scripts/deploy.sh
   ```

`scripts/deploy.sh` will:
1. Build the Linux binary (`GOOS=linux GOARCH=amd64 go build`) and deploy `sample-business-services` to Cloud Run with `--iap` enabled.
2. Create the `SERVICE_ACCOUNT` if it doesn't exist, grant it `roles/run.invoker`, and grant `roles/iam.serviceAccountTokenCreator` to the Apigee service agent (`service-<PROJECT_NUMBER>@gcp-sa-apigee.iam.gserviceaccount.com`).
3. Configure Cloud Run IAP access (`roles/iap.httpsResourceAccessor`) for the service account and `IAP_MEMBERS` (or your active `gcloud` account).
4. Deploy all 12 Apigee proxies (6 REST + 6 MCP) and 12 unlimited-access API products from `apigee/` using `aft`, targeting the deployed Cloud Run URL with Google ID token authentication.

---

## REST API Examples

### 1. Business KPI Summary
```bash
curl -s http://localhost:8080/api/v1/summary | jq
```

### 2. List & Filter Customers / Materials
```bash
# Search customers in Germany
curl -s "http://localhost:8080/api/v1/customers?country=DE" | jq

# Find materials at or below their reorder point
curl -s "http://localhost:8080/api/v1/materials?lowStock=true" | jq
```

### 3. Order-to-Cash Flow (Create Sales Order -> Invoice -> Pay)
```bash
# Create a sales order (checks credit & reserves stock)
curl -s -X POST http://localhost:8080/api/v1/sales-orders \
  -H "Content-Type: application/json" \
  -d '{
    "customerId": "CUST-1001",
    "referencePo": "PO-9981",
    "items": [
      { "materialId": "MAT-1001", "quantity": 2 },
      { "materialId": "MAT-1003", "quantity": 5 }
    ]
  }' | jq

# Generate an invoice from the new sales order (replace SO-50011 with returned ID)
curl -s -X POST http://localhost:8080/api/v1/invoices \
  -H "Content-Type: application/json" \
  -d '{ "salesOrderId": "SO-50011" }' | jq

# Record payment (frees customer credit)
curl -s -X POST http://localhost:8080/api/v1/invoices/INV-90008/pay \
  -H "Content-Type: application/json" \
  -d '{ "paymentReference": "WIRE-2026-01" }' | jq
```

### 4. Procure-to-Pay Flow (Create PO -> Goods Receipt)
```bash
# Receive an open purchase order (increments material stock)
curl -s -X POST http://localhost:8080/api/v1/purchase-orders/PO-45001/receive | jq
```

---

## MCP Endpoint Usage (`/mcp`)

The `/mcp` endpoint implements **MCP Streamable HTTP (JSON-RPC 2.0)**. Changes made via MCP immediately reflect in the REST API and vice versa.

### MCP Client Configuration
Point any HTTP-capable MCP client to:
```json
{
  "mcpServers": {
    "sample-business-service": {
      "type": "http",
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

### Testing `/mcp` with `curl`
```bash
# Initialize session
curl -s -i -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2025-03-26",
      "clientInfo": { "name": "curl", "version": "1.0" }
    }
  }'

# List available MCP tools
curl -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{ "jsonrpc": "2.0", "id": 2, "method": "tools/list" }' | jq

# Call a tool (e.g., find overdue invoices)
curl -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "list_invoices",
      "arguments": { "status": "OVERDUE" }
    }
  }' | jq
```

---

## Project Structure

```text
.
├── main.go                     # Entrypoint, PORT config, graceful shutdown
├── scripts/
│   ├── deploy.sh               # Cloud Run + IAP + Apigee (aft) deployment script
│   └── register_systemd.sh     # Build and register systemd service on Debian Linux
├── env.sample                  # Sample environment variables for deployment
├── public/                     # Static monochrome Explorer UI (light/dark)
├── apigee/                     # 12 Apigee proxy YAMLs (REST & MCP) + 12 product YAMLs
├── internal/
│   ├── domain/
│   │   ├── models.go           # Domain structs (Customer, Material, Vendor, SO, Invoice, PO)
│   │   ├── store.go            # Thread-safe in-memory store & business rules
│   │   └── seed.go             # Pre-loaded demo data + optional ./data/*.json loader
│   ├── rest/
│   │   ├── handlers.go         # /api/v1/* REST handlers
│   │   ├── docs.go             # /openapi.json & /docs Swagger UI
│   │   └── openapi.json        # OpenAPI 3.2.0 specification
│   ├── mcp/
│   │   ├── server.go           # JSON-RPC 2.0 MCP Streamable HTTP server (/mcp)
│   │   ├── schemas.go          # Shared MCP input/output JSON Schemas
│   │   └── tools.go            # MCP tool definitions (inputSchema, outputSchema, structuredContent)
│   └── server/
│       ├── server.go           # HTTP router & middleware (CORS, logging, auth hook)
│       └── server_test.go      # End-to-end REST + OpenAPI + MCP tests
├── go.mod
└── README.md
```

### Adding More Demo Data, New Domains, or Auth
- **Extra Demo Data (no recompile needed):** Drop JSON array files (`customers.json`, `materials.json`, `vendors.json`, `sales_orders.json`, `invoices.json`, `purchase_orders.json`) into `./data` (or set `SEED_DIR=/path/to/dir`) and restart.
- **New Domain:** Add models & store methods in `internal/domain/`, REST handlers in `internal/rest/handlers.go` (plus `openapi.json`), and MCP tools in `internal/mcp/tools.go`. The frontend in `public/` auto-discovers new domains and tools directly from `/openapi.json` and `/mcp`.
- **Authentication:** Implement token/key validation inside `authMiddleware` in `internal/server/server.go`.
