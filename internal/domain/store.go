package domain

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	ErrNotFound     = errors.New("resource not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrConflict     = errors.New("business rule conflict")
)

// Store is a thread-safe in-memory repository for all enterprise business domains.
// Both REST handlers and MCP tools operate on the same Store instance.
type Store struct {
	mu             sync.RWMutex
	customers      map[string]*Customer
	materials      map[string]*Material
	vendors        map[string]*Vendor
	salesOrders    map[string]*SalesOrder
	invoices       map[string]*Invoice
	purchaseOrders map[string]*PurchaseOrder

	nextCustomerSeq int
	nextMaterialSeq int
	nextVendorSeq   int
	nextSOSeq       int
	nextInvoiceSeq  int
	nextPOSeq       int
}

// NewStore initializes a Store pre-populated with realistic enterprise demo data.
func NewStore() *Store {
	s := &Store{
		customers:       make(map[string]*Customer),
		materials:       make(map[string]*Material),
		vendors:         make(map[string]*Vendor),
		salesOrders:     make(map[string]*SalesOrder),
		invoices:        make(map[string]*Invoice),
		purchaseOrders:  make(map[string]*PurchaseOrder),
		nextCustomerSeq: 1001,
		nextMaterialSeq: 1001,
		nextVendorSeq:   1001,
		nextSOSeq:       50001,
		nextInvoiceSeq:  90001,
		nextPOSeq:       45001,
	}
	s.seedDemoData()
	return s
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// --- Customers ---

type CustomerFilter struct {
	Query    string
	Status   string
	Country  string
	SalesOrg string
}

func (s *Store) ListCustomers(f CustomerFilter) []Customer {
	s.mu.RLock()
	defer s.mu.RUnlock()

	q := strings.ToLower(strings.TrimSpace(f.Query))
	out := make([]Customer, 0, len(s.customers))
	for _, c := range s.customers {
		if f.Status != "" && !strings.EqualFold(c.Status, f.Status) {
			continue
		}
		if f.Country != "" && !strings.EqualFold(c.Country, f.Country) {
			continue
		}
		if f.SalesOrg != "" && c.SalesOrg != f.SalesOrg {
			continue
		}
		if q != "" {
			hay := strings.ToLower(c.ID + " " + c.Name + " " + c.LegalName + " " + c.City + " " + c.Industry)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, *c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (s *Store) GetCustomer(id string) (Customer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	c, ok := s.customers[strings.ToUpper(strings.TrimSpace(id))]
	if !ok {
		return Customer{}, fmt.Errorf("%w: customer %s", ErrNotFound, id)
	}
	return *c, nil
}

type CreateCustomerInput struct {
	Name         string  `json:"name"`
	LegalName    string  `json:"legalName"`
	TaxID        string  `json:"taxId"`
	Industry     string  `json:"industry"`
	SalesOrg     string  `json:"salesOrg"`
	Currency     string  `json:"currency"`
	PaymentTerms string  `json:"paymentTerms"`
	CreditLimit  float64 `json:"creditLimit"`
	Country      string  `json:"country"`
	City         string  `json:"city"`
	Street       string  `json:"street"`
	PostalCode   string  `json:"postalCode"`
	Email        string  `json:"email"`
	Phone        string  `json:"phone"`
}

func (s *Store) CreateCustomer(in CreateCustomerInput) (Customer, error) {
	if strings.TrimSpace(in.Name) == "" {
		return Customer{}, fmt.Errorf("%w: name is required", ErrInvalidInput)
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	id := fmt.Sprintf("CUST-%d", s.nextCustomerSeq)
	s.nextCustomerSeq++

	if in.LegalName == "" {
		in.LegalName = in.Name
	}
	if in.SalesOrg == "" {
		in.SalesOrg = "1000"
	}
	if in.Currency == "" {
		in.Currency = "USD"
	}
	if in.PaymentTerms == "" {
		in.PaymentTerms = "NET30"
	}
	if in.CreditLimit <= 0 {
		in.CreditLimit = 50000
	}
	if in.Country == "" {
		in.Country = "US"
	}

	c := &Customer{
		ID:           id,
		Name:         in.Name,
		LegalName:    in.LegalName,
		TaxID:        in.TaxID,
		Industry:     in.Industry,
		SalesOrg:     in.SalesOrg,
		Currency:     strings.ToUpper(in.Currency),
		PaymentTerms: strings.ToUpper(in.PaymentTerms),
		CreditLimit:  round2(in.CreditLimit),
		CreditUsed:   0,
		Status:       "ACTIVE",
		Country:      strings.ToUpper(in.Country),
		City:         in.City,
		Street:       in.Street,
		PostalCode:   in.PostalCode,
		Email:        in.Email,
		Phone:        in.Phone,
		CreatedAt:    time.Now().UTC(),
	}
	s.customers[id] = c
	return *c, nil
}

type UpdateCustomerInput struct {
	Name         *string  `json:"name,omitempty"`
	PaymentTerms *string  `json:"paymentTerms,omitempty"`
	CreditLimit  *float64 `json:"creditLimit,omitempty"`
	Status       *string  `json:"status,omitempty"`
	Email        *string  `json:"email,omitempty"`
	Phone        *string  `json:"phone,omitempty"`
}

func (s *Store) UpdateCustomer(id string, in UpdateCustomerInput) (Customer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, ok := s.customers[strings.ToUpper(strings.TrimSpace(id))]
	if !ok {
		return Customer{}, fmt.Errorf("%w: customer %s", ErrNotFound, id)
	}
	if in.Name != nil && strings.TrimSpace(*in.Name) != "" {
		c.Name = *in.Name
	}
	if in.PaymentTerms != nil && strings.TrimSpace(*in.PaymentTerms) != "" {
		c.PaymentTerms = strings.ToUpper(*in.PaymentTerms)
	}
	if in.CreditLimit != nil {
		if *in.CreditLimit < 0 {
			return Customer{}, fmt.Errorf("%w: creditLimit cannot be negative", ErrInvalidInput)
		}
		c.CreditLimit = round2(*in.CreditLimit)
	}
	if in.Status != nil {
		st := strings.ToUpper(strings.TrimSpace(*in.Status))
		if st != "ACTIVE" && st != "BLOCKED" && st != "INACTIVE" {
			return Customer{}, fmt.Errorf("%w: status must be ACTIVE, BLOCKED, or INACTIVE", ErrInvalidInput)
		}
		c.Status = st
	}
	if in.Email != nil {
		c.Email = *in.Email
	}
	if in.Phone != nil {
		c.Phone = *in.Phone
	}
	return *c, nil
}

// --- Materials ---

type MaterialFilter struct {
	Query    string
	Type     string
	Category string
	LowStock bool
}

func (s *Store) ListMaterials(f MaterialFilter) []Material {
	s.mu.RLock()
	defer s.mu.RUnlock()

	q := strings.ToLower(strings.TrimSpace(f.Query))
	out := make([]Material, 0, len(s.materials))
	for _, m := range s.materials {
		if f.Type != "" && !strings.EqualFold(m.Type, f.Type) {
			continue
		}
		if f.Category != "" && !strings.EqualFold(m.Category, f.Category) {
			continue
		}
		if f.LowStock && m.StockQty > m.ReorderPoint {
			continue
		}
		if q != "" {
			hay := strings.ToLower(m.ID + " " + m.SKU + " " + m.Name + " " + m.Description + " " + m.Category)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, *m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (s *Store) GetMaterial(id string) (Material, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	m, ok := s.materials[strings.ToUpper(strings.TrimSpace(id))]
	if !ok {
		return Material{}, fmt.Errorf("%w: material %s", ErrNotFound, id)
	}
	return *m, nil
}

// --- Vendors ---

type VendorFilter struct {
	Query   string
	Status  string
	Country string
}

func (s *Store) ListVendors(f VendorFilter) []Vendor {
	s.mu.RLock()
	defer s.mu.RUnlock()

	q := strings.ToLower(strings.TrimSpace(f.Query))
	out := make([]Vendor, 0, len(s.vendors))
	for _, v := range s.vendors {
		if f.Status != "" && !strings.EqualFold(v.Status, f.Status) {
			continue
		}
		if f.Country != "" && !strings.EqualFold(v.Country, f.Country) {
			continue
		}
		if q != "" {
			hay := strings.ToLower(v.ID + " " + v.Name + " " + v.Category + " " + v.City)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, *v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (s *Store) GetVendor(id string) (Vendor, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.vendors[strings.ToUpper(strings.TrimSpace(id))]
	if !ok {
		return Vendor{}, fmt.Errorf("%w: vendor %s", ErrNotFound, id)
	}
	return *v, nil
}

// --- Sales Orders ---

type SalesOrderFilter struct {
	CustomerID string
	Status     string
	SalesOrg   string
}

func (s *Store) ListSalesOrders(f SalesOrderFilter) []SalesOrder {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]SalesOrder, 0, len(s.salesOrders))
	for _, so := range s.salesOrders {
		if f.CustomerID != "" && !strings.EqualFold(so.CustomerID, f.CustomerID) {
			continue
		}
		if f.Status != "" && !strings.EqualFold(so.Status, f.Status) {
			continue
		}
		if f.SalesOrg != "" && so.SalesOrg != f.SalesOrg {
			continue
		}
		cp := *so
		cp.Items = append([]SalesOrderItem(nil), so.Items...)
		out = append(out, cp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	return out
}

func (s *Store) GetSalesOrder(id string) (SalesOrder, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	so, ok := s.salesOrders[strings.ToUpper(strings.TrimSpace(id))]
	if !ok {
		return SalesOrder{}, fmt.Errorf("%w: sales order %s", ErrNotFound, id)
	}
	cp := *so
	cp.Items = append([]SalesOrderItem(nil), so.Items...)
	return cp, nil
}

type CreateSalesOrderItemInput struct {
	MaterialID string `json:"materialId"`
	Quantity   int    `json:"quantity"`
}

type CreateSalesOrderInput struct {
	CustomerID            string                      `json:"customerId"`
	RequestedDeliveryDate string                      `json:"requestedDeliveryDate"`
	ReferencePO           string                      `json:"referencePo"`
	Notes                 string                      `json:"notes"`
	Items                 []CreateSalesOrderItemInput `json:"items"`
}

func (s *Store) CreateSalesOrder(in CreateSalesOrderInput) (SalesOrder, error) {
	if strings.TrimSpace(in.CustomerID) == "" {
		return SalesOrder{}, fmt.Errorf("%w: customerId is required", ErrInvalidInput)
	}
	if len(in.Items) == 0 {
		return SalesOrder{}, fmt.Errorf("%w: at least one line item is required", ErrInvalidInput)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	cust, ok := s.customers[strings.ToUpper(strings.TrimSpace(in.CustomerID))]
	if !ok {
		return SalesOrder{}, fmt.Errorf("%w: customer %s", ErrNotFound, in.CustomerID)
	}
	if cust.Status == "BLOCKED" {
		return SalesOrder{}, fmt.Errorf("%w: customer %s is credit-blocked", ErrConflict, cust.ID)
	}

	items := make([]SalesOrderItem, 0, len(in.Items))
	var netTotal float64

	// Validate all materials and stock before mutating state
	type stockDeduction struct {
		mat *Material
		qty int
	}
	var deductions []stockDeduction

	for idx, it := range in.Items {
		if it.Quantity <= 0 {
			return SalesOrder{}, fmt.Errorf("%w: item %d quantity must be > 0", ErrInvalidInput, idx+1)
		}
		mat, ok := s.materials[strings.ToUpper(strings.TrimSpace(it.MaterialID))]
		if !ok {
			return SalesOrder{}, fmt.Errorf("%w: material %s", ErrNotFound, it.MaterialID)
		}
		if mat.StockQty < it.Quantity {
			return SalesOrder{}, fmt.Errorf("%w: insufficient stock for %s (%s): available %d, requested %d",
				ErrConflict, mat.ID, mat.Name, mat.StockQty, it.Quantity)
		}
		lineNet := round2(mat.UnitPrice * float64(it.Quantity))
		netTotal += lineNet
		items = append(items, SalesOrderItem{
			ItemNumber:   (idx + 1) * 10,
			MaterialID:   mat.ID,
			MaterialName: mat.Name,
			Quantity:     it.Quantity,
			UOM:          mat.BaseUOM,
			UnitPrice:    mat.UnitPrice,
			NetAmount:    lineNet,
		})
		deductions = append(deductions, stockDeduction{mat: mat, qty: it.Quantity})
	}

	netTotal = round2(netTotal)
	tax := round2(netTotal * 0.08)
	grossTotal := round2(netTotal + tax)

	if cust.CreditUsed+grossTotal > cust.CreditLimit {
		return SalesOrder{}, fmt.Errorf("%w: order total %.2f %s exceeds remaining credit limit (used %.2f / limit %.2f)",
			ErrConflict, grossTotal, cust.Currency, cust.CreditUsed, cust.CreditLimit)
	}

	// Commit stock & credit mutations
	for _, d := range deductions {
		d.mat.StockQty -= d.qty
	}
	cust.CreditUsed = round2(cust.CreditUsed + grossTotal)

	id := fmt.Sprintf("SO-%d", s.nextSOSeq)
	s.nextSOSeq++

	now := time.Now().UTC()
	reqDeliv := in.RequestedDeliveryDate
	if reqDeliv == "" {
		reqDeliv = now.AddDate(0, 0, 7).Format("2006-01-02")
	}

	so := &SalesOrder{
		ID:                    id,
		CustomerID:            cust.ID,
		CustomerName:          cust.Name,
		SalesOrg:              cust.SalesOrg,
		OrderDate:             now.Format("2006-01-02"),
		RequestedDeliveryDate: reqDeliv,
		Status:                "OPEN",
		Currency:              cust.Currency,
		NetAmount:             netTotal,
		TaxAmount:             tax,
		TotalAmount:           grossTotal,
		ReferencePO:           in.ReferencePO,
		Notes:                 in.Notes,
		Items:                 items,
		CreatedAt:             now,
	}
	s.salesOrders[id] = so

	cp := *so
	cp.Items = append([]SalesOrderItem(nil), so.Items...)
	return cp, nil
}

func (s *Store) UpdateSalesOrderStatus(id string, status string) (SalesOrder, error) {
	st := strings.ToUpper(strings.TrimSpace(status))
	switch st {
	case "OPEN", "CONFIRMED", "DELIVERED", "BILLED", "CANCELLED":
	default:
		return SalesOrder{}, fmt.Errorf("%w: invalid status %q", ErrInvalidInput, status)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	so, ok := s.salesOrders[strings.ToUpper(strings.TrimSpace(id))]
	if !ok {
		return SalesOrder{}, fmt.Errorf("%w: sales order %s", ErrNotFound, id)
	}
	if so.Status == "CANCELLED" && st != "CANCELLED" {
		return SalesOrder{}, fmt.Errorf("%w: cannot reopen cancelled sales order %s", ErrConflict, so.ID)
	}
	if st == "CANCELLED" && so.Status != "CANCELLED" {
		// Release reserved stock and customer credit
		for _, it := range so.Items {
			if m, ok := s.materials[it.MaterialID]; ok {
				m.StockQty += it.Quantity
			}
		}
		if c, ok := s.customers[so.CustomerID]; ok {
			c.CreditUsed = round2(math.Max(0, c.CreditUsed-so.TotalAmount))
		}
	}
	so.Status = st

	cp := *so
	cp.Items = append([]SalesOrderItem(nil), so.Items...)
	return cp, nil
}

// --- Invoices / Billing ---

type InvoiceFilter struct {
	CustomerID   string
	SalesOrderID string
	Status       string
}

func (s *Store) ListInvoices(f InvoiceFilter) []Invoice {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Invoice, 0, len(s.invoices))
	for _, inv := range s.invoices {
		if f.CustomerID != "" && !strings.EqualFold(inv.CustomerID, f.CustomerID) {
			continue
		}
		if f.SalesOrderID != "" && !strings.EqualFold(inv.SalesOrderID, f.SalesOrderID) {
			continue
		}
		if f.Status != "" && !strings.EqualFold(inv.Status, f.Status) {
			continue
		}
		cp := *inv
		cp.Items = append([]InvoiceItem(nil), inv.Items...)
		out = append(out, cp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	return out
}

func (s *Store) GetInvoice(id string) (Invoice, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	inv, ok := s.invoices[strings.ToUpper(strings.TrimSpace(id))]
	if !ok {
		return Invoice{}, fmt.Errorf("%w: invoice %s", ErrNotFound, id)
	}
	cp := *inv
	cp.Items = append([]InvoiceItem(nil), inv.Items...)
	return cp, nil
}

type CreateInvoiceInput struct {
	SalesOrderID string `json:"salesOrderId"`
	DueDays      int    `json:"dueDays"`
}

func (s *Store) CreateInvoiceFromSalesOrder(in CreateInvoiceInput) (Invoice, error) {
	if strings.TrimSpace(in.SalesOrderID) == "" {
		return Invoice{}, fmt.Errorf("%w: salesOrderId is required", ErrInvalidInput)
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	so, ok := s.salesOrders[strings.ToUpper(strings.TrimSpace(in.SalesOrderID))]
	if !ok {
		return Invoice{}, fmt.Errorf("%w: sales order %s", ErrNotFound, in.SalesOrderID)
	}
	if so.Status == "CANCELLED" {
		return Invoice{}, fmt.Errorf("%w: cannot bill cancelled sales order %s", ErrConflict, so.ID)
	}

	dueDays := in.DueDays
	if dueDays <= 0 {
		dueDays = 30
	}

	items := make([]InvoiceItem, 0, len(so.Items))
	for _, it := range so.Items {
		items = append(items, InvoiceItem{
			ItemNumber:  it.ItemNumber,
			MaterialID:  it.MaterialID,
			Description: it.MaterialName,
			Quantity:    it.Quantity,
			UOM:         it.UOM,
			UnitPrice:   it.UnitPrice,
			NetAmount:   it.NetAmount,
		})
	}

	now := time.Now().UTC()
	id := fmt.Sprintf("INV-%d", s.nextInvoiceSeq)
	s.nextInvoiceSeq++

	inv := &Invoice{
		ID:           id,
		SalesOrderID: so.ID,
		CustomerID:   so.CustomerID,
		CustomerName: so.CustomerName,
		BillingDate:  now.Format("2006-01-02"),
		DueDate:      now.AddDate(0, 0, dueDays).Format("2006-01-02"),
		Status:       "OPEN",
		Currency:     so.Currency,
		NetAmount:    so.NetAmount,
		TaxAmount:    so.TaxAmount,
		TotalAmount:  so.TotalAmount,
		PaidAmount:   0,
		Items:        items,
		CreatedAt:    now,
	}
	s.invoices[id] = inv
	so.Status = "BILLED"

	cp := *inv
	cp.Items = append([]InvoiceItem(nil), inv.Items...)
	return cp, nil
}

type PayInvoiceInput struct {
	Amount           float64 `json:"amount"`
	PaymentReference string  `json:"paymentReference"`
}

func (s *Store) PayInvoice(id string, in PayInvoiceInput) (Invoice, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	inv, ok := s.invoices[strings.ToUpper(strings.TrimSpace(id))]
	if !ok {
		return Invoice{}, fmt.Errorf("%w: invoice %s", ErrNotFound, id)
	}
	if inv.Status == "PAID" {
		return Invoice{}, fmt.Errorf("%w: invoice %s is already paid", ErrConflict, inv.ID)
	}
	if inv.Status == "CANCELLED" {
		return Invoice{}, fmt.Errorf("%w: cannot pay cancelled invoice %s", ErrConflict, inv.ID)
	}

	remaining := round2(inv.TotalAmount - inv.PaidAmount)
	payAmt := in.Amount
	if payAmt <= 0 {
		payAmt = remaining
	}
	if payAmt > remaining+0.01 {
		return Invoice{}, fmt.Errorf("%w: payment amount %.2f exceeds remaining balance %.2f", ErrInvalidInput, payAmt, remaining)
	}

	inv.PaidAmount = round2(inv.PaidAmount + payAmt)
	if in.PaymentReference != "" {
		inv.PaymentReference = in.PaymentReference
	} else if inv.PaymentReference == "" {
		inv.PaymentReference = fmt.Sprintf("PAY-%d", time.Now().Unix())
	}

	if inv.PaidAmount >= inv.TotalAmount-0.01 {
		inv.PaidAmount = inv.TotalAmount
		inv.Status = "PAID"
		now := time.Now().UTC()
		inv.PaidAt = &now
	}

	if c, ok := s.customers[inv.CustomerID]; ok {
		c.CreditUsed = round2(math.Max(0, c.CreditUsed-payAmt))
	}

	cp := *inv
	cp.Items = append([]InvoiceItem(nil), inv.Items...)
	return cp, nil
}

// --- Purchase Orders / Procurement ---

type PurchaseOrderFilter struct {
	VendorID string
	Status   string
	Plant    string
}

func (s *Store) ListPurchaseOrders(f PurchaseOrderFilter) []PurchaseOrder {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]PurchaseOrder, 0, len(s.purchaseOrders))
	for _, po := range s.purchaseOrders {
		if f.VendorID != "" && !strings.EqualFold(po.VendorID, f.VendorID) {
			continue
		}
		if f.Status != "" && !strings.EqualFold(po.Status, f.Status) {
			continue
		}
		if f.Plant != "" && po.Plant != f.Plant {
			continue
		}
		cp := *po
		cp.Items = append([]PurchaseOrderItem(nil), po.Items...)
		out = append(out, cp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	return out
}

func (s *Store) GetPurchaseOrder(id string) (PurchaseOrder, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	po, ok := s.purchaseOrders[strings.ToUpper(strings.TrimSpace(id))]
	if !ok {
		return PurchaseOrder{}, fmt.Errorf("%w: purchase order %s", ErrNotFound, id)
	}
	cp := *po
	cp.Items = append([]PurchaseOrderItem(nil), po.Items...)
	return cp, nil
}

type CreatePurchaseOrderItemInput struct {
	MaterialID string  `json:"materialId"`
	Quantity   int     `json:"quantity"`
	UnitPrice  float64 `json:"unitPrice,omitempty"`
}

type CreatePurchaseOrderInput struct {
	VendorID             string                         `json:"vendorId"`
	Plant                string                         `json:"plant"`
	ExpectedDeliveryDate string                         `json:"expectedDeliveryDate"`
	Notes                string                         `json:"notes"`
	Items                []CreatePurchaseOrderItemInput `json:"items"`
}

func (s *Store) CreatePurchaseOrder(in CreatePurchaseOrderInput) (PurchaseOrder, error) {
	if strings.TrimSpace(in.VendorID) == "" {
		return PurchaseOrder{}, fmt.Errorf("%w: vendorId is required", ErrInvalidInput)
	}
	if len(in.Items) == 0 {
		return PurchaseOrder{}, fmt.Errorf("%w: at least one line item is required", ErrInvalidInput)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	vend, ok := s.vendors[strings.ToUpper(strings.TrimSpace(in.VendorID))]
	if !ok {
		return PurchaseOrder{}, fmt.Errorf("%w: vendor %s", ErrNotFound, in.VendorID)
	}
	if vend.Status == "ON_HOLD" {
		return PurchaseOrder{}, fmt.Errorf("%w: vendor %s is ON_HOLD", ErrConflict, vend.ID)
	}

	items := make([]PurchaseOrderItem, 0, len(in.Items))
	var total float64
	for idx, it := range in.Items {
		if it.Quantity <= 0 {
			return PurchaseOrder{}, fmt.Errorf("%w: item %d quantity must be > 0", ErrInvalidInput, idx+1)
		}
		mat, ok := s.materials[strings.ToUpper(strings.TrimSpace(it.MaterialID))]
		if !ok {
			return PurchaseOrder{}, fmt.Errorf("%w: material %s", ErrNotFound, it.MaterialID)
		}
		unitPrice := it.UnitPrice
		if unitPrice <= 0 {
			unitPrice = round2(mat.UnitPrice * 0.65) // default wholesale procurement cost
		}
		lineNet := round2(unitPrice * float64(it.Quantity))
		total += lineNet
		items = append(items, PurchaseOrderItem{
			ItemNumber:   (idx + 1) * 10,
			MaterialID:   mat.ID,
			MaterialName: mat.Name,
			Quantity:     it.Quantity,
			UOM:          mat.BaseUOM,
			UnitPrice:    unitPrice,
			NetAmount:    lineNet,
		})
	}

	plant := in.Plant
	if plant == "" {
		plant = "1000"
	}
	now := time.Now().UTC()
	expDeliv := in.ExpectedDeliveryDate
	if expDeliv == "" {
		expDeliv = now.AddDate(0, 0, 14).Format("2006-01-02")
	}

	id := fmt.Sprintf("PO-%d", s.nextPOSeq)
	s.nextPOSeq++

	po := &PurchaseOrder{
		ID:                   id,
		VendorID:             vend.ID,
		VendorName:           vend.Name,
		PurchasingOrg:        "1000",
		Plant:                plant,
		OrderDate:            now.Format("2006-01-02"),
		ExpectedDeliveryDate: expDeliv,
		Status:               "ORDERED",
		Currency:             vend.Currency,
		TotalAmount:          round2(total),
		Notes:                in.Notes,
		Items:                items,
		CreatedAt:            now,
	}
	s.purchaseOrders[id] = po

	cp := *po
	cp.Items = append([]PurchaseOrderItem(nil), po.Items...)
	return cp, nil
}

// ReceivePurchaseOrder posts a Goods Receipt (MIGO) against a PO, incrementing material stock.
func (s *Store) ReceivePurchaseOrder(id string) (PurchaseOrder, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	po, ok := s.purchaseOrders[strings.ToUpper(strings.TrimSpace(id))]
	if !ok {
		return PurchaseOrder{}, fmt.Errorf("%w: purchase order %s", ErrNotFound, id)
	}
	if po.Status == "RECEIVED" {
		return PurchaseOrder{}, fmt.Errorf("%w: purchase order %s is already received", ErrConflict, po.ID)
	}
	if po.Status == "CANCELLED" {
		return PurchaseOrder{}, fmt.Errorf("%w: cannot receive cancelled purchase order %s", ErrConflict, po.ID)
	}

	for _, it := range po.Items {
		if mat, ok := s.materials[it.MaterialID]; ok {
			mat.StockQty += it.Quantity
		}
	}
	po.Status = "RECEIVED"

	cp := *po
	cp.Items = append([]PurchaseOrderItem(nil), po.Items...)
	return cp, nil
}

// --- Business Summary ---

func (s *Store) GetBusinessSummary() BusinessSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var sum BusinessSummary
	sum.TotalCustomers = len(s.customers)
	for _, c := range s.customers {
		if c.Status == "ACTIVE" {
			sum.ActiveCustomers++
		}
	}

	sum.TotalMaterials = len(s.materials)
	for _, m := range s.materials {
		if m.StockQty <= m.ReorderPoint {
			sum.LowStockMaterials++
		}
	}

	sum.TotalVendors = len(s.vendors)

	sum.TotalSalesOrders = len(s.salesOrders)
	for _, so := range s.salesOrders {
		if so.Status == "OPEN" || so.Status == "CONFIRMED" {
			sum.OpenSalesOrders++
		}
	}

	sum.TotalInvoices = len(s.invoices)
	for _, inv := range s.invoices {
		if inv.Status != "CANCELLED" {
			sum.TotalBilledAmount = round2(sum.TotalBilledAmount + inv.TotalAmount)
		}
		if inv.Status == "OPEN" {
			sum.OpenInvoices++
			sum.OutstandingARAmount = round2(sum.OutstandingARAmount + (inv.TotalAmount - inv.PaidAmount))
		} else if inv.Status == "OVERDUE" {
			sum.OverdueInvoices++
			sum.OutstandingARAmount = round2(sum.OutstandingARAmount + (inv.TotalAmount - inv.PaidAmount))
		}
	}

	sum.TotalPurchaseOrders = len(s.purchaseOrders)
	for _, po := range s.purchaseOrders {
		if po.Status == "DRAFT" || po.Status == "APPROVED" || po.Status == "ORDERED" {
			sum.OpenPurchaseOrders++
			sum.OpenPOCommittedSpend = round2(sum.OpenPOCommittedSpend + po.TotalAmount)
		}
	}

	return sum
}
