package mcp

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tyayers/sample-business-services/internal/domain"
	"github.com/tyayers/sample-business-services/internal/rest"
)

const protocolVersion = "2025-03-26"

// Server implements an MCP (Model Context Protocol) Streamable HTTP endpoint (/mcp)
// backed by the shared domain.Store.
type Server struct {
	store *domain.Store
	tools map[string]toolDef
	order []string
}

type toolDef struct {
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	InputSchema  map[string]any `json:"inputSchema"`
	OutputSchema map[string]any `json:"outputSchema,omitempty"`
	Handler      func(args json.RawMessage) (any, error)
}

func NewServer(store *domain.Store) *Server {
	s := &Server{
		store: store,
		tools: make(map[string]toolDef),
	}
	s.registerTools()
	return s
}

func (s *Server) addTool(t toolDef) {
	s.tools[t.Name] = t
	s.order = append(s.order, t.Name)
}

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /mcp", s.handlePost)
	mux.HandleFunc("DELETE /mcp", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /mcp", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "method_not_allowed",
			"message": "Use POST /mcp with JSON-RPC 2.0 requests (MCP Streamable HTTP transport).",
		})
	})
}

// --- JSON-RPC 2.0 types ---

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (s *Server) handlePost(w http.ResponseWriter, r *http.Request) {
	var req rpcRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeRPCError(w, nil, -32700, "parse error: "+err.Error())
		return
	}
	if req.JSONRPC != "" && req.JSONRPC != "2.0" {
		writeRPCError(w, req.ID, -32600, "invalid jsonrpc version")
		return
	}

	// Notifications have no ID; acknowledge with 202 Accepted.
	if len(req.ID) == 0 || string(req.ID) == "null" {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	switch req.Method {
	case "initialize":
		sid := r.Header.Get("Mcp-Session-Id")
		if sid == "" {
			sid = newSessionID()
		}
		w.Header().Set("Mcp-Session-Id", sid)
		writeRPCResult(w, req.ID, map[string]any{
			"protocolVersion": protocolVersion,
			"capabilities": map[string]any{
				"tools":     map[string]any{"listChanged": false},
				"resources": map[string]any{"subscribe": false, "listChanged": false},
			},
			"serverInfo": map[string]any{
				"name":    "sample-business-service",
				"version": "1.0.0",
			},
		})

	case "ping":
		writeRPCResult(w, req.ID, map[string]any{})

	case "tools/list":
		list := make([]map[string]any, 0, len(s.order))
		for _, name := range s.order {
			t := s.tools[name]
			entry := map[string]any{
				"name":        t.Name,
				"description": t.Description,
				"inputSchema": t.InputSchema,
			}
			if t.OutputSchema != nil {
				entry["outputSchema"] = t.OutputSchema
			}
			list = append(list, entry)
		}
		writeRPCResult(w, req.ID, map[string]any{"tools": list})

	case "tools/call":
		var params struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			writeRPCError(w, req.ID, -32602, "invalid params: "+err.Error())
			return
		}
		t, ok := s.tools[params.Name]
		if !ok {
			writeRPCError(w, req.ID, -32602, fmt.Sprintf("unknown tool: %s", params.Name))
			return
		}
		args := params.Arguments
		if len(args) == 0 {
			args = json.RawMessage("{}")
		}
		res, err := t.Handler(args)
		if err != nil {
			writeRPCResult(w, req.ID, map[string]any{
				"content": []map[string]any{{"type": "text", "text": err.Error()}},
				"isError": true,
			})
			return
		}
		writeRPCResult(w, req.ID, map[string]any{
			"structuredContent": res,
			"isError":           false,
		})

	case "resources/list":
		writeRPCResult(w, req.ID, map[string]any{
			"resources": []map[string]any{
				{
					"uri":         "business://summary",
					"name":        "Business KPI Summary",
					"description": "Live cross-domain KPI snapshot across customers, materials, sales orders, invoices, and purchase orders.",
					"mimeType":    "application/json",
				},
				{
					"uri":         "business://openapi.json",
					"name":        "OpenAPI 3.2.0 Specification",
					"description": "Complete OpenAPI specification for the REST business services.",
					"mimeType":    "application/json",
				},
			},
		})

	case "resources/read":
		var params struct {
			URI string `json:"uri"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			writeRPCError(w, req.ID, -32602, "invalid params: "+err.Error())
			return
		}
		switch params.URI {
		case "business://summary":
			b, _ := json.MarshalIndent(s.store.GetBusinessSummary(), "", "  ")
			writeRPCResult(w, req.ID, map[string]any{
				"contents": []map[string]any{{"uri": params.URI, "mimeType": "application/json", "text": string(b)}},
			})
		case "business://openapi.json":
			writeRPCResult(w, req.ID, map[string]any{
				"contents": []map[string]any{{"uri": params.URI, "mimeType": "application/json", "text": string(rest.OpenAPISpec)}},
			})
		default:
			writeRPCError(w, req.ID, -32602, fmt.Sprintf("unknown resource uri: %s", params.URI))
		}

	default:
		writeRPCError(w, req.ID, -32601, fmt.Sprintf("method not found: %s", req.Method))
	}
}

func writeRPCResult(w http.ResponseWriter, id json.RawMessage, result any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(rpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	})
}

func writeRPCError(w http.ResponseWriter, id json.RawMessage, code int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(rpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &rpcError{Code: code, Message: msg},
	})
}

func newSessionID() string {
	var b [12]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
