package domain

import "time"

// Customer represents a Business Partner / Customer Master record.
type Customer struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	LegalName    string    `json:"legalName"`
	TaxID        string    `json:"taxId"`
	Industry     string    `json:"industry"`
	SalesOrg     string    `json:"salesOrg"`
	Currency     string    `json:"currency"`
	PaymentTerms string    `json:"paymentTerms"`
	CreditLimit  float64   `json:"creditLimit"`
	CreditUsed   float64   `json:"creditUsed"`
	Status       string    `json:"status"` // ACTIVE, BLOCKED, INACTIVE
	Country      string    `json:"country"`
	City         string    `json:"city"`
	Street       string    `json:"street"`
	PostalCode   string    `json:"postalCode"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	CreatedAt    time.Time `json:"createdAt"`
}

// Material represents a Material Master record.
type Material struct {
	ID              string  `json:"id"`
	SKU             string  `json:"sku"`
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	Type            string  `json:"type"` // FERT (Finished), HAWA (Trading), ROH (Raw)
	Category        string  `json:"category"`
	BaseUOM         string  `json:"baseUom"`
	UnitPrice       float64 `json:"unitPrice"`
	Currency        string  `json:"currency"`
	StockQty        int     `json:"stockQty"`
	ReorderPoint    int     `json:"reorderPoint"`
	Plant           string  `json:"plant"`
	StorageLocation string  `json:"storageLocation"`
	Status          string  `json:"status"` // ACTIVE, DISCONTINUED
}

// Vendor represents a Supplier / Vendor Master record.
type Vendor struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Category     string `json:"category"`
	Country      string `json:"country"`
	City         string `json:"city"`
	Currency     string `json:"currency"`
	PaymentTerms string `json:"paymentTerms"`
	Rating       string `json:"rating"` // A, B, C
	Email        string `json:"email"`
	TaxID        string `json:"taxId"`
	Status       string `json:"status"` // ACTIVE, ON_HOLD
}

// SalesOrderItem represents a line item on a Sales Order (VBAP).
type SalesOrderItem struct {
	ItemNumber   int     `json:"itemNumber"`
	MaterialID   string  `json:"materialId"`
	MaterialName string  `json:"materialName"`
	Quantity     int     `json:"quantity"`
	UOM          string  `json:"uom"`
	UnitPrice    float64 `json:"unitPrice"`
	NetAmount    float64 `json:"netAmount"`
}

// SalesOrder represents a Sales Order header.
type SalesOrder struct {
	ID                    string           `json:"id"`
	CustomerID            string           `json:"customerId"`
	CustomerName          string           `json:"customerName"`
	SalesOrg              string           `json:"salesOrg"`
	OrderDate             string           `json:"orderDate"`
	RequestedDeliveryDate string           `json:"requestedDeliveryDate"`
	Status                string           `json:"status"` // OPEN, CONFIRMED, DELIVERED, BILLED, CANCELLED
	Currency              string           `json:"currency"`
	NetAmount             float64          `json:"netAmount"`
	TaxAmount             float64          `json:"taxAmount"`
	TotalAmount           float64          `json:"totalAmount"`
	ReferencePO           string           `json:"referencePo,omitempty"`
	Notes                 string           `json:"notes,omitempty"`
	Items                 []SalesOrderItem `json:"items"`
	CreatedAt             time.Time        `json:"createdAt"`
}

// InvoiceItem represents a billing document line item (VBRP).
type InvoiceItem struct {
	ItemNumber  int     `json:"itemNumber"`
	MaterialID  string  `json:"materialId"`
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	UOM         string  `json:"uom"`
	UnitPrice   float64 `json:"unitPrice"`
	NetAmount   float64 `json:"netAmount"`
}

// Invoice represents a Billing Document / Invoice.
type Invoice struct {
	ID               string        `json:"id"`
	SalesOrderID     string        `json:"salesOrderId,omitempty"`
	CustomerID       string        `json:"customerId"`
	CustomerName     string        `json:"customerName"`
	BillingDate      string        `json:"billingDate"`
	DueDate          string        `json:"dueDate"`
	Status           string        `json:"status"` // OPEN, PAID, OVERDUE, CANCELLED
	Currency         string        `json:"currency"`
	NetAmount        float64       `json:"netAmount"`
	TaxAmount        float64       `json:"taxAmount"`
	TotalAmount      float64       `json:"totalAmount"`
	PaidAmount       float64       `json:"paidAmount"`
	PaymentReference string        `json:"paymentReference,omitempty"`
	PaidAt           *time.Time    `json:"paidAt,omitempty"`
	Items            []InvoiceItem `json:"items"`
	CreatedAt        time.Time     `json:"createdAt"`
}

// PurchaseOrderItem represents a line item on a Purchase Order (EKPO).
type PurchaseOrderItem struct {
	ItemNumber   int     `json:"itemNumber"`
	MaterialID   string  `json:"materialId"`
	MaterialName string  `json:"materialName"`
	Quantity     int     `json:"quantity"`
	UOM          string  `json:"uom"`
	UnitPrice    float64 `json:"unitPrice"`
	NetAmount    float64 `json:"netAmount"`
}

// PurchaseOrder represents a Purchase Order header.
type PurchaseOrder struct {
	ID                   string              `json:"id"`
	VendorID             string              `json:"vendorId"`
	VendorName           string              `json:"vendorName"`
	PurchasingOrg        string              `json:"purchasingOrg"`
	Plant                string              `json:"plant"`
	OrderDate            string              `json:"orderDate"`
	ExpectedDeliveryDate string              `json:"expectedDeliveryDate"`
	Status               string              `json:"status"` // DRAFT, APPROVED, ORDERED, RECEIVED, CANCELLED
	Currency             string              `json:"currency"`
	TotalAmount          float64             `json:"totalAmount"`
	Notes                string              `json:"notes,omitempty"`
	Items                []PurchaseOrderItem `json:"items"`
	CreatedAt            time.Time           `json:"createdAt"`
}

// BusinessSummary provides a cross-domain KPI snapshot.
type BusinessSummary struct {
	TotalCustomers       int     `json:"totalCustomers"`
	ActiveCustomers      int     `json:"activeCustomers"`
	TotalMaterials       int     `json:"totalMaterials"`
	LowStockMaterials    int     `json:"lowStockMaterials"`
	TotalVendors         int     `json:"totalVendors"`
	TotalSalesOrders     int     `json:"totalSalesOrders"`
	OpenSalesOrders      int     `json:"openSalesOrders"`
	TotalInvoices        int     `json:"totalInvoices"`
	OpenInvoices         int     `json:"openInvoices"`
	OverdueInvoices      int     `json:"overdueInvoices"`
	TotalBilledAmount    float64 `json:"totalBilledAmount"`
	OutstandingARAmount  float64 `json:"outstandingArAmount"`
	TotalPurchaseOrders  int     `json:"totalPurchaseOrders"`
	OpenPurchaseOrders   int     `json:"openPurchaseOrders"`
	OpenPOCommittedSpend float64 `json:"openPoCommittedSpend"`
}
