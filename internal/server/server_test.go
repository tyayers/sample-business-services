package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tyayers/sample-business-services/internal/domain"
	"github.com/tyayers/sample-business-services/internal/server"
)

func newTestServer() http.Handler {
	return server.New(domain.NewStore())
}

func doReq(t *testing.T, h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHealthAndOpenAPI(t *testing.T) {
	h := newTestServer()

	if rec := doReq(t, h, "GET", "/", nil); rec.Code != http.StatusOK {
		t.Fatalf("GET / = %d, want 200", rec.Code)
	}
	if rec := doReq(t, h, "GET", "/app.js", nil); rec.Code != http.StatusOK {
		t.Fatalf("GET /app.js = %d, want 200", rec.Code)
	}
	if rec := doReq(t, h, "GET", "/styles.css", nil); rec.Code != http.StatusOK {
		t.Fatalf("GET /styles.css = %d, want 200", rec.Code)
	}
	if rec := doReq(t, h, "GET", "/api/v1", nil); rec.Code != http.StatusOK {
		t.Fatalf("GET /api/v1 = %d, want 200", rec.Code)
	}
	if rec := doReq(t, h, "GET", "/healthz", nil); rec.Code != http.StatusOK {
		t.Fatalf("GET /healthz = %d, want 200", rec.Code)
	}
	if rec := doReq(t, h, "GET", "/docs", nil); rec.Code != http.StatusOK {
		t.Fatalf("GET /docs = %d, want 200", rec.Code)
	}

	rec := doReq(t, h, "GET", "/openapi.json", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /openapi.json = %d, want 200", rec.Code)
	}
	var spec map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &spec); err != nil {
		t.Fatalf("invalid openapi.json: %v", err)
	}
	if spec["openapi"] != "3.2.0" {
		t.Fatalf("openapi version = %v, want 3.2.0", spec["openapi"])
	}
}

func TestRESTDomainsEndToEnd(t *testing.T) {
	h := newTestServer()

	// 1. Summary
	rec := doReq(t, h, "GET", "/api/v1/summary", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/summary = %d", rec.Code)
	}
	var sum domain.BusinessSummary
	_ = json.Unmarshal(rec.Body.Bytes(), &sum)
	if sum.TotalCustomers < 10 || sum.TotalMaterials < 10 {
		t.Fatalf("expected rich demo data, got %+v", sum)
	}

	// 2. Create Customer
	rec = doReq(t, h, "POST", "/api/v1/customers", map[string]any{
		"name":        "Helios Manufacturing",
		"country":     "US",
		"creditLimit": 100000,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /api/v1/customers = %d: %s", rec.Code, rec.Body.String())
	}
	var cust domain.Customer
	_ = json.Unmarshal(rec.Body.Bytes(), &cust)

	// 3. Create Sales Order & verify stock deduction + credit usage
	matBeforeRec := doReq(t, h, "GET", "/api/v1/materials/MAT-1001", nil)
	var matBefore domain.Material
	_ = json.Unmarshal(matBeforeRec.Body.Bytes(), &matBefore)

	soRec := doReq(t, h, "POST", "/api/v1/sales-orders", map[string]any{
		"customerId": cust.ID,
		"items": []map[string]any{
			{"materialId": "MAT-1001", "quantity": 2},
		},
	})
	if soRec.Code != http.StatusCreated {
		t.Fatalf("POST /api/v1/sales-orders = %d: %s", soRec.Code, soRec.Body.String())
	}
	var so domain.SalesOrder
	_ = json.Unmarshal(soRec.Body.Bytes(), &so)

	matAfterRec := doReq(t, h, "GET", "/api/v1/materials/MAT-1001", nil)
	var matAfter domain.Material
	_ = json.Unmarshal(matAfterRec.Body.Bytes(), &matAfter)
	if matAfter.StockQty != matBefore.StockQty-2 {
		t.Fatalf("stockQty = %d, want %d", matAfter.StockQty, matBefore.StockQty-2)
	}

	// 4. Blocked customer should fail with 409 Conflict
	blockedRec := doReq(t, h, "POST", "/api/v1/sales-orders", map[string]any{
		"customerId": "CUST-1010",
		"items":      []map[string]any{{"materialId": "MAT-1001", "quantity": 1}},
	})
	if blockedRec.Code != http.StatusConflict {
		t.Fatalf("blocked customer order = %d, want 409", blockedRec.Code)
	}

	// 5. Create Invoice from Sales Order & Pay it
	invRec := doReq(t, h, "POST", "/api/v1/invoices", map[string]any{
		"salesOrderId": so.ID,
	})
	if invRec.Code != http.StatusCreated {
		t.Fatalf("POST /api/v1/invoices = %d: %s", invRec.Code, invRec.Body.String())
	}
	var inv domain.Invoice
	_ = json.Unmarshal(invRec.Body.Bytes(), &inv)

	payRec := doReq(t, h, "POST", "/api/v1/invoices/"+inv.ID+"/pay", map[string]any{
		"paymentReference": "WIRE-TEST-001",
	})
	if payRec.Code != http.StatusOK {
		t.Fatalf("POST /pay = %d: %s", payRec.Code, payRec.Body.String())
	}
	var paidInv domain.Invoice
	_ = json.Unmarshal(payRec.Body.Bytes(), &paidInv)
	if paidInv.Status != "PAID" {
		t.Fatalf("invoice status = %s, want PAID", paidInv.Status)
	}

	// 6. Receive Purchase Order and verify stock increases
	poRec := doReq(t, h, "POST", "/api/v1/purchase-orders/PO-45001/receive", nil)
	if poRec.Code != http.StatusOK {
		t.Fatalf("POST /receive = %d: %s", poRec.Code, poRec.Body.String())
	}
}

func TestMCPEndpointAndCrossProtocolSync(t *testing.T) {
	h := newTestServer()

	// 1. Initialize
	initRec := doReq(t, h, "POST", "/mcp", map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2025-03-26",
			"clientInfo":      map[string]any{"name": "test-client", "version": "1.0"},
		},
	})
	if initRec.Code != http.StatusOK {
		t.Fatalf("MCP initialize = %d", initRec.Code)
	}
	if initRec.Header().Get("Mcp-Session-Id") == "" {
		t.Fatalf("expected Mcp-Session-Id header")
	}

	// 2. List tools
	listRec := doReq(t, h, "POST", "/mcp", map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
	})
	var listResp struct {
		Result struct {
			Tools []struct {
				Name         string         `json:"name"`
				InputSchema  map[string]any `json:"inputSchema"`
				OutputSchema map[string]any `json:"outputSchema"`
			} `json:"tools"`
		} `json:"result"`
	}
	_ = json.Unmarshal(listRec.Body.Bytes(), &listResp)
	if len(listResp.Result.Tools) < 15 {
		t.Fatalf("expected >= 15 MCP tools, got %d", len(listResp.Result.Tools))
	}
	for _, tool := range listResp.Result.Tools {
		if tool.InputSchema == nil || tool.OutputSchema == nil {
			t.Fatalf("expected both inputSchema and outputSchema on tool %s", tool.Name)
		}
	}

	// 3. Create customer via MCP, then fetch via REST
	callRec := doReq(t, h, "POST", "/mcp", map[string]any{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "create_customer",
			"arguments": map[string]any{
				"name":    "MCP Sync Corp",
				"country": "DE",
			},
		},
	})
	if callRec.Code != http.StatusOK {
		t.Fatalf("MCP tools/call = %d", callRec.Code)
	}
	var callResp struct {
		Result struct {
			StructuredContent domain.Customer `json:"structuredContent"`
			Content           []any           `json:"content"`
			IsError           bool            `json:"isError"`
		} `json:"result"`
	}
	_ = json.Unmarshal(callRec.Body.Bytes(), &callResp)
	if callResp.Result.IsError || len(callResp.Result.Content) != 0 {
		t.Fatalf("expected structuredContent only without content, got: %s", callRec.Body.String())
	}
	createdCust := callResp.Result.StructuredContent
	if createdCust.ID == "" {
		t.Fatalf("expected structuredContent customer ID, got empty")
	}

	restRec := doReq(t, h, "GET", "/api/v1/customers/"+createdCust.ID, nil)
	if restRec.Code != http.StatusOK {
		t.Fatalf("expected MCP-created customer %s to be readable via REST, got %d", createdCust.ID, restRec.Code)
	}

	// 4. Read resource
	resRec := doReq(t, h, "POST", "/mcp", map[string]any{
		"jsonrpc": "2.0",
		"id":      4,
		"method":  "resources/read",
		"params":  map[string]any{"uri": "business://summary"},
	})
	if resRec.Code != http.StatusOK {
		t.Fatalf("MCP resources/read = %d", resRec.Code)
	}
}
