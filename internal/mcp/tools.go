package mcp

import (
	"encoding/json"

	"github.com/tyayers/sample-business-services/internal/domain"
)

func (s *Server) registerTools() {
	// --- Summary ---
	s.addTool(toolDef{
		Name:         "get_business_summary",
		Description:  "Get a cross-domain KPI snapshot across customers, materials, sales orders, invoices, and purchase orders.",
		InputSchema:  objSchema(map[string]any{}),
		OutputSchema: summarySchema,
		Handler: func(args json.RawMessage) (any, error) {
			return s.store.GetBusinessSummary(), nil
		},
	})

	// --- Customers ---
	s.addTool(toolDef{
		Name:        "list_customers",
		Description: "List or search Customer Master records. Supports filtering by query text, status (ACTIVE/BLOCKED/INACTIVE), country, and salesOrg.",
		InputSchema: objSchema(map[string]any{
			"q":        map[string]any{"type": "string", "description": "Search text matching ID, name, city, or industry"},
			"status":   map[string]any{"type": "string", "enum": []string{"ACTIVE", "BLOCKED", "INACTIVE"}},
			"country":  map[string]any{"type": "string", "description": "2-letter country code, e.g. US, DE, JP"},
			"salesOrg": map[string]any{"type": "string"},
		}),
		OutputSchema: listEnvelopeSchema(customerSchema),
		Handler: func(args json.RawMessage) (any, error) {
			var in struct {
				Q        string `json:"q"`
				Status   string `json:"status"`
				Country  string `json:"country"`
				SalesOrg string `json:"salesOrg"`
			}
			if err := json.Unmarshal(args, &in); err != nil {
				return nil, err
			}
			items := s.store.ListCustomers(domain.CustomerFilter{
				Query: in.Q, Status: in.Status, Country: in.Country, SalesOrg: in.SalesOrg,
			})
			return map[string]any{"items": items, "count": len(items)}, nil
		},
	})

	s.addTool(toolDef{
		Name:        "get_customer",
		Description: "Get full details for a customer by ID (e.g. CUST-1001), including credit limit and credit used.",
		InputSchema: objSchema(map[string]any{
			"id": map[string]any{"type": "string", "description": "Customer ID (e.g. CUST-1001)"},
		}, "id"),
		OutputSchema: customerSchema,
		Handler: func(args json.RawMessage) (any, error) {
			var in struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(args, &in); err != nil {
				return nil, err
			}
			return s.store.GetCustomer(in.ID)
		},
	})

	s.addTool(toolDef{
		Name:        "create_customer",
		Description: "Create a new Customer Master record.",
		InputSchema: objSchema(map[string]any{
			"name":         map[string]any{"type": "string"},
			"legalName":    map[string]any{"type": "string"},
			"industry":     map[string]any{"type": "string"},
			"country":      map[string]any{"type": "string", "default": "US"},
			"city":         map[string]any{"type": "string"},
			"currency":     map[string]any{"type": "string", "default": "USD"},
			"paymentTerms": map[string]any{"type": "string", "default": "NET30"},
			"creditLimit":  map[string]any{"type": "number", "default": 50000},
			"email":        map[string]any{"type": "string"},
		}, "name"),
		OutputSchema: customerSchema,
		Handler: func(args json.RawMessage) (any, error) {
			var in domain.CreateCustomerInput
			if err := json.Unmarshal(args, &in); err != nil {
				return nil, err
			}
			return s.store.CreateCustomer(in)
		},
	})

	s.addTool(toolDef{
		Name:        "update_customer",
		Description: "Update customer status (ACTIVE/BLOCKED/INACTIVE), creditLimit, paymentTerms, or contact info.",
		InputSchema: objSchema(map[string]any{
			"id":           map[string]any{"type": "string"},
			"status":       map[string]any{"type": "string", "enum": []string{"ACTIVE", "BLOCKED", "INACTIVE"}},
			"creditLimit":  map[string]any{"type": "number"},
			"paymentTerms": map[string]any{"type": "string"},
			"email":        map[string]any{"type": "string"},
		}, "id"),
		OutputSchema: customerSchema,
		Handler: func(args json.RawMessage) (any, error) {
			var in struct {
				ID           string   `json:"id"`
				Status       *string  `json:"status"`
				CreditLimit  *float64 `json:"creditLimit"`
				PaymentTerms *string  `json:"paymentTerms"`
				Email        *string  `json:"email"`
			}
			if err := json.Unmarshal(args, &in); err != nil {
				return nil, err
			}
			return s.store.UpdateCustomer(in.ID, domain.UpdateCustomerInput{
				Status: in.Status, CreditLimit: in.CreditLimit, PaymentTerms: in.PaymentTerms, Email: in.Email,
			})
		},
	})

	// --- Materials ---
	s.addTool(toolDef{
		Name:        "list_materials",
		Description: "List materials/products and inventory levels. Set lowStock=true to find items at or below their reorder point.",
		InputSchema: objSchema(map[string]any{
			"q":        map[string]any{"type": "string", "description": "Search by ID, SKU, name, or category"},
			"type":     map[string]any{"type": "string", "enum": []string{"FERT", "HAWA", "ROH"}},
			"category": map[string]any{"type": "string"},
			"lowStock": map[string]any{"type": "boolean"},
		}),
		OutputSchema: listEnvelopeSchema(materialSchema),
		Handler: func(args json.RawMessage) (any, error) {
			var in struct {
				Q        string `json:"q"`
				Type     string `json:"type"`
				Category string `json:"category"`
				LowStock bool   `json:"lowStock"`
			}
			if err := json.Unmarshal(args, &in); err != nil {
				return nil, err
			}
			items := s.store.ListMaterials(domain.MaterialFilter{
				Query: in.Q, Type: in.Type, Category: in.Category, LowStock: in.LowStock,
			})
			return map[string]any{"items": items, "count": len(items)}, nil
		},
	})

	s.addTool(toolDef{
		Name:        "get_material",
		Description: "Get material details and stock level by ID (e.g. MAT-1001).",
		InputSchema: objSchema(map[string]any{
			"id": map[string]any{"type": "string"},
		}, "id"),
		OutputSchema: materialSchema,
		Handler: func(args json.RawMessage) (any, error) {
			var in struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(args, &in); err != nil {
				return nil, err
			}
			return s.store.GetMaterial(in.ID)
		},
	})

	// --- Vendors ---
	s.addTool(toolDef{
		Name:        "list_vendors",
		Description: "List suppliers/vendors. Supports filtering by query, status (ACTIVE/ON_HOLD), and country.",
		InputSchema: objSchema(map[string]any{
			"q":       map[string]any{"type": "string"},
			"status":  map[string]any{"type": "string", "enum": []string{"ACTIVE", "ON_HOLD"}},
			"country": map[string]any{"type": "string"},
		}),
		OutputSchema: listEnvelopeSchema(vendorSchema),
		Handler: func(args json.RawMessage) (any, error) {
			var in struct {
				Q       string `json:"q"`
				Status  string `json:"status"`
				Country string `json:"country"`
			}
			if err := json.Unmarshal(args, &in); err != nil {
				return nil, err
			}
			items := s.store.ListVendors(domain.VendorFilter{Query: in.Q, Status: in.Status, Country: in.Country})
			return map[string]any{"items": items, "count": len(items)}, nil
		},
	})

	s.addTool(toolDef{
		Name:        "get_vendor",
		Description: "Get vendor details by ID (e.g. VEND-1001).",
		InputSchema: objSchema(map[string]any{
			"id": map[string]any{"type": "string"},
		}, "id"),
		OutputSchema: vendorSchema,
		Handler: func(args json.RawMessage) (any, error) {
			var in struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(args, &in); err != nil {
				return nil, err
			}
			return s.store.GetVendor(in.ID)
		},
	})

	// --- Sales Orders ---
	s.addTool(toolDef{
		Name:        "list_sales_orders",
		Description: "List sales orders, optionally filtered by customerId, status, or salesOrg.",
		InputSchema: objSchema(map[string]any{
			"customerId": map[string]any{"type": "string"},
			"status":     map[string]any{"type": "string", "enum": []string{"OPEN", "CONFIRMED", "DELIVERED", "BILLED", "CANCELLED"}},
			"salesOrg":   map[string]any{"type": "string"},
		}),
		OutputSchema: listEnvelopeSchema(salesOrderSchema),
		Handler: func(args json.RawMessage) (any, error) {
			var in struct {
				CustomerID string `json:"customerId"`
				Status     string `json:"status"`
				SalesOrg   string `json:"salesOrg"`
			}
			if err := json.Unmarshal(args, &in); err != nil {
				return nil, err
			}
			items := s.store.ListSalesOrders(domain.SalesOrderFilter{
				CustomerID: in.CustomerID, Status: in.Status, SalesOrg: in.SalesOrg,
			})
			return map[string]any{"items": items, "count": len(items)}, nil
		},
	})

	s.addTool(toolDef{
		Name:        "get_sales_order",
		Description: "Get a sales order and its line items by ID (e.g. SO-50001).",
		InputSchema: objSchema(map[string]any{
			"id": map[string]any{"type": "string"},
		}, "id"),
		OutputSchema: salesOrderSchema,
		Handler: func(args json.RawMessage) (any, error) {
			var in struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(args, &in); err != nil {
				return nil, err
			}
			return s.store.GetSalesOrder(in.ID)
		},
	})

	s.addTool(toolDef{
		Name:        "create_sales_order",
		Description: "Create a new sales order. Validates customer credit limit and material stock, reserving both.",
		InputSchema: objSchema(map[string]any{
			"customerId":            map[string]any{"type": "string", "description": "Customer ID (e.g. CUST-1001)"},
			"requestedDeliveryDate": map[string]any{"type": "string", "description": "YYYY-MM-DD"},
			"referencePo":           map[string]any{"type": "string"},
			"notes":                 map[string]any{"type": "string"},
			"items": map[string]any{
				"type": "array",
				"items": objSchema(map[string]any{
					"materialId": map[string]any{"type": "string"},
					"quantity":   map[string]any{"type": "integer"},
				}, "materialId", "quantity"),
			},
		}, "customerId", "items"),
		OutputSchema: salesOrderSchema,
		Handler: func(args json.RawMessage) (any, error) {
			var in domain.CreateSalesOrderInput
			if err := json.Unmarshal(args, &in); err != nil {
				return nil, err
			}
			return s.store.CreateSalesOrder(in)
		},
	})

	s.addTool(toolDef{
		Name:        "update_sales_order_status",
		Description: "Update a sales order status. Setting CANCELLED releases reserved stock and customer credit.",
		InputSchema: objSchema(map[string]any{
			"id":     map[string]any{"type": "string"},
			"status": map[string]any{"type": "string", "enum": []string{"OPEN", "CONFIRMED", "DELIVERED", "BILLED", "CANCELLED"}},
		}, "id", "status"),
		OutputSchema: salesOrderSchema,
		Handler: func(args json.RawMessage) (any, error) {
			var in struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			}
			if err := json.Unmarshal(args, &in); err != nil {
				return nil, err
			}
			return s.store.UpdateSalesOrderStatus(in.ID, in.Status)
		},
	})

	// --- Invoices / Billing ---
	s.addTool(toolDef{
		Name:        "list_invoices",
		Description: "List billing invoices, optionally filtered by customerId, salesOrderId, or status (OPEN/PAID/OVERDUE/CANCELLED).",
		InputSchema: objSchema(map[string]any{
			"customerId":   map[string]any{"type": "string"},
			"salesOrderId": map[string]any{"type": "string"},
			"status":       map[string]any{"type": "string", "enum": []string{"OPEN", "PAID", "OVERDUE", "CANCELLED"}},
		}),
		OutputSchema: listEnvelopeSchema(invoiceSchema),
		Handler: func(args json.RawMessage) (any, error) {
			var in struct {
				CustomerID   string `json:"customerId"`
				SalesOrderID string `json:"salesOrderId"`
				Status       string `json:"status"`
			}
			if err := json.Unmarshal(args, &in); err != nil {
				return nil, err
			}
			items := s.store.ListInvoices(domain.InvoiceFilter{
				CustomerID: in.CustomerID, SalesOrderID: in.SalesOrderID, Status: in.Status,
			})
			return map[string]any{"items": items, "count": len(items)}, nil
		},
	})

	s.addTool(toolDef{
		Name:        "get_invoice",
		Description: "Get invoice details by ID (e.g. INV-90001).",
		InputSchema: objSchema(map[string]any{
			"id": map[string]any{"type": "string"},
		}, "id"),
		OutputSchema: invoiceSchema,
		Handler: func(args json.RawMessage) (any, error) {
			var in struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(args, &in); err != nil {
				return nil, err
			}
			return s.store.GetInvoice(in.ID)
		},
	})

	s.addTool(toolDef{
		Name:        "create_invoice",
		Description: "Create a billing invoice from an existing sales order and mark the sales order BILLED.",
		InputSchema: objSchema(map[string]any{
			"salesOrderId": map[string]any{"type": "string"},
			"dueDays":      map[string]any{"type": "integer", "default": 30},
		}, "salesOrderId"),
		OutputSchema: invoiceSchema,
		Handler: func(args json.RawMessage) (any, error) {
			var in domain.CreateInvoiceInput
			if err := json.Unmarshal(args, &in); err != nil {
				return nil, err
			}
			return s.store.CreateInvoiceFromSalesOrder(in)
		},
	})

	s.addTool(toolDef{
		Name:        "pay_invoice",
		Description: "Record a full or partial payment against an invoice and free up customer credit.",
		InputSchema: objSchema(map[string]any{
			"id":               map[string]any{"type": "string"},
			"amount":           map[string]any{"type": "number", "description": "Omit or 0 to pay full remaining balance"},
			"paymentReference": map[string]any{"type": "string"},
		}, "id"),
		OutputSchema: invoiceSchema,
		Handler: func(args json.RawMessage) (any, error) {
			var in struct {
				ID               string  `json:"id"`
				Amount           float64 `json:"amount"`
				PaymentReference string  `json:"paymentReference"`
			}
			if err := json.Unmarshal(args, &in); err != nil {
				return nil, err
			}
			return s.store.PayInvoice(in.ID, domain.PayInvoiceInput{
				Amount: in.Amount, PaymentReference: in.PaymentReference,
			})
		},
	})

	// --- Purchase Orders / Procurement ---
	s.addTool(toolDef{
		Name:        "list_purchase_orders",
		Description: "List procurement purchase orders, optionally filtered by vendorId, status, or plant.",
		InputSchema: objSchema(map[string]any{
			"vendorId": map[string]any{"type": "string"},
			"status":   map[string]any{"type": "string", "enum": []string{"DRAFT", "APPROVED", "ORDERED", "RECEIVED", "CANCELLED"}},
			"plant":    map[string]any{"type": "string"},
		}),
		OutputSchema: listEnvelopeSchema(purchaseOrderSchema),
		Handler: func(args json.RawMessage) (any, error) {
			var in struct {
				VendorID string `json:"vendorId"`
				Status   string `json:"status"`
				Plant    string `json:"plant"`
			}
			if err := json.Unmarshal(args, &in); err != nil {
				return nil, err
			}
			items := s.store.ListPurchaseOrders(domain.PurchaseOrderFilter{
				VendorID: in.VendorID, Status: in.Status, Plant: in.Plant,
			})
			return map[string]any{"items": items, "count": len(items)}, nil
		},
	})

	s.addTool(toolDef{
		Name:        "get_purchase_order",
		Description: "Get purchase order details by ID (e.g. PO-45001).",
		InputSchema: objSchema(map[string]any{
			"id": map[string]any{"type": "string"},
		}, "id"),
		OutputSchema: purchaseOrderSchema,
		Handler: func(args json.RawMessage) (any, error) {
			var in struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(args, &in); err != nil {
				return nil, err
			}
			return s.store.GetPurchaseOrder(in.ID)
		},
	})

	s.addTool(toolDef{
		Name:        "create_purchase_order",
		Description: "Create a new purchase order with a vendor for one or more materials.",
		InputSchema: objSchema(map[string]any{
			"vendorId":             map[string]any{"type": "string"},
			"plant":                map[string]any{"type": "string", "default": "1000"},
			"expectedDeliveryDate": map[string]any{"type": "string"},
			"notes":                map[string]any{"type": "string"},
			"items": map[string]any{
				"type": "array",
				"items": objSchema(map[string]any{
					"materialId": map[string]any{"type": "string"},
					"quantity":   map[string]any{"type": "integer"},
					"unitPrice":  map[string]any{"type": "number"},
				}, "materialId", "quantity"),
			},
		}, "vendorId", "items"),
		OutputSchema: purchaseOrderSchema,
		Handler: func(args json.RawMessage) (any, error) {
			var in domain.CreatePurchaseOrderInput
			if err := json.Unmarshal(args, &in); err != nil {
				return nil, err
			}
			return s.store.CreatePurchaseOrder(in)
		},
	})

	s.addTool(toolDef{
		Name:        "receive_purchase_order",
		Description: "Post a goods receipt (MIGO) for a purchase order, marking it RECEIVED and incrementing material stock.",
		InputSchema: objSchema(map[string]any{
			"id": map[string]any{"type": "string"},
		}, "id"),
		OutputSchema: purchaseOrderSchema,
		Handler: func(args json.RawMessage) (any, error) {
			var in struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(args, &in); err != nil {
				return nil, err
			}
			return s.store.ReceivePurchaseOrder(in.ID)
		},
	})
}
