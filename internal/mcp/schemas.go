package mcp

func objSchema(props map[string]any, required ...string) map[string]any {
	s := map[string]any{
		"type":       "object",
		"properties": props,
	}
	if len(required) > 0 {
		s["required"] = required
	}
	return s
}

func listEnvelopeSchema(itemSchema map[string]any) map[string]any {
	return objSchema(map[string]any{
		"items": map[string]any{"type": "array", "items": itemSchema},
		"count": map[string]any{"type": "integer"},
	}, "items", "count")
}

var (
	summarySchema = objSchema(map[string]any{
		"totalCustomers":       map[string]any{"type": "integer"},
		"activeCustomers":      map[string]any{"type": "integer"},
		"totalMaterials":       map[string]any{"type": "integer"},
		"lowStockMaterials":    map[string]any{"type": "integer"},
		"totalVendors":         map[string]any{"type": "integer"},
		"totalSalesOrders":     map[string]any{"type": "integer"},
		"openSalesOrders":      map[string]any{"type": "integer"},
		"totalInvoices":        map[string]any{"type": "integer"},
		"openInvoices":         map[string]any{"type": "integer"},
		"overdueInvoices":      map[string]any{"type": "integer"},
		"totalBilledAmount":    map[string]any{"type": "number"},
		"outstandingArAmount":  map[string]any{"type": "number"},
		"totalPurchaseOrders":  map[string]any{"type": "integer"},
		"openPurchaseOrders":   map[string]any{"type": "integer"},
		"openPoCommittedSpend": map[string]any{"type": "number"},
	})

	customerSchema = objSchema(map[string]any{
		"id":           map[string]any{"type": "string"},
		"name":         map[string]any{"type": "string"},
		"legalName":    map[string]any{"type": "string"},
		"taxId":        map[string]any{"type": "string"},
		"industry":     map[string]any{"type": "string"},
		"salesOrg":     map[string]any{"type": "string"},
		"currency":     map[string]any{"type": "string"},
		"paymentTerms": map[string]any{"type": "string"},
		"creditLimit":  map[string]any{"type": "number"},
		"creditUsed":   map[string]any{"type": "number"},
		"status":       map[string]any{"type": "string", "enum": []string{"ACTIVE", "BLOCKED", "INACTIVE"}},
		"country":      map[string]any{"type": "string"},
		"city":         map[string]any{"type": "string"},
		"street":       map[string]any{"type": "string"},
		"postalCode":   map[string]any{"type": "string"},
		"email":        map[string]any{"type": "string"},
		"phone":        map[string]any{"type": "string"},
		"createdAt":    map[string]any{"type": "string", "format": "date-time"},
	}, "id", "name", "status")

	materialSchema = objSchema(map[string]any{
		"id":              map[string]any{"type": "string"},
		"sku":             map[string]any{"type": "string"},
		"name":            map[string]any{"type": "string"},
		"description":     map[string]any{"type": "string"},
		"type":            map[string]any{"type": "string", "enum": []string{"FERT", "HAWA", "ROH"}},
		"category":        map[string]any{"type": "string"},
		"baseUom":         map[string]any{"type": "string"},
		"unitPrice":       map[string]any{"type": "number"},
		"currency":        map[string]any{"type": "string"},
		"stockQty":        map[string]any{"type": "integer"},
		"reorderPoint":    map[string]any{"type": "integer"},
		"plant":           map[string]any{"type": "string"},
		"storageLocation": map[string]any{"type": "string"},
		"status":          map[string]any{"type": "string"},
	}, "id", "sku", "name", "stockQty")

	vendorSchema = objSchema(map[string]any{
		"id":           map[string]any{"type": "string"},
		"name":         map[string]any{"type": "string"},
		"category":     map[string]any{"type": "string"},
		"country":      map[string]any{"type": "string"},
		"city":         map[string]any{"type": "string"},
		"currency":     map[string]any{"type": "string"},
		"paymentTerms": map[string]any{"type": "string"},
		"rating":       map[string]any{"type": "string"},
		"email":        map[string]any{"type": "string"},
		"taxId":        map[string]any{"type": "string"},
		"status":       map[string]any{"type": "string", "enum": []string{"ACTIVE", "ON_HOLD"}},
	}, "id", "name", "status")

	salesOrderItemSchema = objSchema(map[string]any{
		"itemNumber":   map[string]any{"type": "integer"},
		"materialId":   map[string]any{"type": "string"},
		"materialName": map[string]any{"type": "string"},
		"quantity":     map[string]any{"type": "integer"},
		"uom":          map[string]any{"type": "string"},
		"unitPrice":    map[string]any{"type": "number"},
		"netAmount":    map[string]any{"type": "number"},
	})

	salesOrderSchema = objSchema(map[string]any{
		"id":                    map[string]any{"type": "string"},
		"customerId":            map[string]any{"type": "string"},
		"customerName":          map[string]any{"type": "string"},
		"salesOrg":              map[string]any{"type": "string"},
		"orderDate":             map[string]any{"type": "string"},
		"requestedDeliveryDate": map[string]any{"type": "string"},
		"status":                map[string]any{"type": "string", "enum": []string{"OPEN", "CONFIRMED", "DELIVERED", "BILLED", "CANCELLED"}},
		"currency":              map[string]any{"type": "string"},
		"netAmount":             map[string]any{"type": "number"},
		"taxAmount":             map[string]any{"type": "number"},
		"totalAmount":           map[string]any{"type": "number"},
		"referencePo":           map[string]any{"type": "string"},
		"notes":                 map[string]any{"type": "string"},
		"items":                 map[string]any{"type": "array", "items": salesOrderItemSchema},
		"createdAt":             map[string]any{"type": "string", "format": "date-time"},
	}, "id", "customerId", "status", "totalAmount")

	invoiceItemSchema = objSchema(map[string]any{
		"itemNumber":  map[string]any{"type": "integer"},
		"materialId":  map[string]any{"type": "string"},
		"description": map[string]any{"type": "string"},
		"quantity":    map[string]any{"type": "integer"},
		"uom":         map[string]any{"type": "string"},
		"unitPrice":   map[string]any{"type": "number"},
		"netAmount":   map[string]any{"type": "number"},
	})

	invoiceSchema = objSchema(map[string]any{
		"id":               map[string]any{"type": "string"},
		"salesOrderId":     map[string]any{"type": "string"},
		"customerId":       map[string]any{"type": "string"},
		"customerName":     map[string]any{"type": "string"},
		"billingDate":      map[string]any{"type": "string"},
		"dueDate":          map[string]any{"type": "string"},
		"status":           map[string]any{"type": "string", "enum": []string{"OPEN", "PAID", "OVERDUE", "CANCELLED"}},
		"currency":         map[string]any{"type": "string"},
		"netAmount":        map[string]any{"type": "number"},
		"taxAmount":        map[string]any{"type": "number"},
		"totalAmount":      map[string]any{"type": "number"},
		"paidAmount":       map[string]any{"type": "number"},
		"paymentReference": map[string]any{"type": "string"},
		"paidAt":           map[string]any{"type": "string", "format": "date-time"},
		"items":            map[string]any{"type": "array", "items": invoiceItemSchema},
		"createdAt":        map[string]any{"type": "string", "format": "date-time"},
	}, "id", "customerId", "status", "totalAmount")

	purchaseOrderItemSchema = objSchema(map[string]any{
		"itemNumber":   map[string]any{"type": "integer"},
		"materialId":   map[string]any{"type": "string"},
		"materialName": map[string]any{"type": "string"},
		"quantity":     map[string]any{"type": "integer"},
		"uom":          map[string]any{"type": "string"},
		"unitPrice":    map[string]any{"type": "number"},
		"netAmount":    map[string]any{"type": "number"},
	})

	purchaseOrderSchema = objSchema(map[string]any{
		"id":                   map[string]any{"type": "string"},
		"vendorId":             map[string]any{"type": "string"},
		"vendorName":           map[string]any{"type": "string"},
		"purchasingOrg":        map[string]any{"type": "string"},
		"plant":                map[string]any{"type": "string"},
		"orderDate":            map[string]any{"type": "string"},
		"expectedDeliveryDate": map[string]any{"type": "string"},
		"status":               map[string]any{"type": "string", "enum": []string{"DRAFT", "APPROVED", "ORDERED", "RECEIVED", "CANCELLED"}},
		"currency":             map[string]any{"type": "string"},
		"totalAmount":          map[string]any{"type": "number"},
		"notes":                map[string]any{"type": "string"},
		"items":                map[string]any{"type": "array", "items": purchaseOrderItemSchema},
		"createdAt":            map[string]any{"type": "string", "format": "date-time"},
	}, "id", "vendorId", "status", "totalAmount")
)
