package rest

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/tyayers/sample-business-services/internal/domain"
)

// Handler exposes the business domains as clean RESTful JSON endpoints.
type Handler struct {
	store *domain.Store
}

func NewHandler(store *domain.Store) *Handler {
	return &Handler{store: store}
}

// RegisterRoutes mounts all /api/v1/* endpoints on the provided ServeMux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/summary", h.getSummary)

	// Customer Master
	mux.HandleFunc("GET /api/v1/customers", h.listCustomers)
	mux.HandleFunc("POST /api/v1/customers", h.createCustomer)
	mux.HandleFunc("GET /api/v1/customers/{id}", h.getCustomer)
	mux.HandleFunc("PATCH /api/v1/customers/{id}", h.updateCustomer)

	// Material Master
	mux.HandleFunc("GET /api/v1/materials", h.listMaterials)
	mux.HandleFunc("GET /api/v1/materials/{id}", h.getMaterial)

	// Vendor Master
	mux.HandleFunc("GET /api/v1/vendors", h.listVendors)
	mux.HandleFunc("GET /api/v1/vendors/{id}", h.getVendor)

	// Sales Orders (SD)
	mux.HandleFunc("GET /api/v1/sales-orders", h.listSalesOrders)
	mux.HandleFunc("POST /api/v1/sales-orders", h.createSalesOrder)
	mux.HandleFunc("GET /api/v1/sales-orders/{id}", h.getSalesOrder)
	mux.HandleFunc("PATCH /api/v1/sales-orders/{id}/status", h.updateSalesOrderStatus)

	// Billing & Invoices (FI/SD)
	mux.HandleFunc("GET /api/v1/invoices", h.listInvoices)
	mux.HandleFunc("POST /api/v1/invoices", h.createInvoice)
	mux.HandleFunc("GET /api/v1/invoices/{id}", h.getInvoice)
	mux.HandleFunc("POST /api/v1/invoices/{id}/pay", h.payInvoice)

	// Procurement / Purchase Orders (MM)
	mux.HandleFunc("GET /api/v1/purchase-orders", h.listPurchaseOrders)
	mux.HandleFunc("POST /api/v1/purchase-orders", h.createPurchaseOrder)
	mux.HandleFunc("GET /api/v1/purchase-orders/{id}", h.getPurchaseOrder)
	mux.HandleFunc("POST /api/v1/purchase-orders/{id}/receive", h.receivePurchaseOrder)
}

type listResponse[T any] struct {
	Items []T `json:"items"`
	Count int `json:"count"`
}

type apiError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeJSON(w, http.StatusNotFound, apiError{Error: "not_found", Message: err.Error()})
	case errors.Is(err, domain.ErrInvalidInput):
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid_input", Message: err.Error()})
	case errors.Is(err, domain.ErrConflict):
		writeJSON(w, http.StatusConflict, apiError{Error: "conflict", Message: err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "internal_error", Message: err.Error()})
	}
}

func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	return nil
}

// --- Summary ---

func (h *Handler) getSummary(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.store.GetBusinessSummary())
}

// --- Customers ---

func (h *Handler) listCustomers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	items := h.store.ListCustomers(domain.CustomerFilter{
		Query:    q.Get("q"),
		Status:   q.Get("status"),
		Country:  q.Get("country"),
		SalesOrg: q.Get("salesOrg"),
	})
	writeJSON(w, http.StatusOK, listResponse[domain.Customer]{Items: items, Count: len(items)})
}

func (h *Handler) getCustomer(w http.ResponseWriter, r *http.Request) {
	c, err := h.store.GetCustomer(r.PathValue("id"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (h *Handler) createCustomer(w http.ResponseWriter, r *http.Request) {
	var in domain.CreateCustomerInput
	if err := decodeJSON(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "bad_request", Message: err.Error()})
		return
	}
	c, err := h.store.CreateCustomer(in)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func (h *Handler) updateCustomer(w http.ResponseWriter, r *http.Request) {
	var in domain.UpdateCustomerInput
	if err := decodeJSON(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "bad_request", Message: err.Error()})
		return
	}
	c, err := h.store.UpdateCustomer(r.PathValue("id"), in)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

// --- Materials ---

func (h *Handler) listMaterials(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	lowStock, _ := strconv.ParseBool(q.Get("lowStock"))
	items := h.store.ListMaterials(domain.MaterialFilter{
		Query:    q.Get("q"),
		Type:     q.Get("type"),
		Category: q.Get("category"),
		LowStock: lowStock,
	})
	writeJSON(w, http.StatusOK, listResponse[domain.Material]{Items: items, Count: len(items)})
}

func (h *Handler) getMaterial(w http.ResponseWriter, r *http.Request) {
	m, err := h.store.GetMaterial(r.PathValue("id"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, m)
}

// --- Vendors ---

func (h *Handler) listVendors(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	items := h.store.ListVendors(domain.VendorFilter{
		Query:   q.Get("q"),
		Status:  q.Get("status"),
		Country: q.Get("country"),
	})
	writeJSON(w, http.StatusOK, listResponse[domain.Vendor]{Items: items, Count: len(items)})
}

func (h *Handler) getVendor(w http.ResponseWriter, r *http.Request) {
	v, err := h.store.GetVendor(r.PathValue("id"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

// --- Sales Orders ---

func (h *Handler) listSalesOrders(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	items := h.store.ListSalesOrders(domain.SalesOrderFilter{
		CustomerID: q.Get("customerId"),
		Status:     q.Get("status"),
		SalesOrg:   q.Get("salesOrg"),
	})
	writeJSON(w, http.StatusOK, listResponse[domain.SalesOrder]{Items: items, Count: len(items)})
}

func (h *Handler) getSalesOrder(w http.ResponseWriter, r *http.Request) {
	so, err := h.store.GetSalesOrder(r.PathValue("id"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, so)
}

func (h *Handler) createSalesOrder(w http.ResponseWriter, r *http.Request) {
	var in domain.CreateSalesOrderInput
	if err := decodeJSON(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "bad_request", Message: err.Error()})
		return
	}
	so, err := h.store.CreateSalesOrder(in)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, so)
}

func (h *Handler) updateSalesOrderStatus(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status string `json:"status"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "bad_request", Message: err.Error()})
		return
	}
	so, err := h.store.UpdateSalesOrderStatus(r.PathValue("id"), body.Status)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, so)
}

// --- Invoices ---

func (h *Handler) listInvoices(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	items := h.store.ListInvoices(domain.InvoiceFilter{
		CustomerID:   q.Get("customerId"),
		SalesOrderID: q.Get("salesOrderId"),
		Status:       q.Get("status"),
	})
	writeJSON(w, http.StatusOK, listResponse[domain.Invoice]{Items: items, Count: len(items)})
}

func (h *Handler) getInvoice(w http.ResponseWriter, r *http.Request) {
	inv, err := h.store.GetInvoice(r.PathValue("id"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, inv)
}

func (h *Handler) createInvoice(w http.ResponseWriter, r *http.Request) {
	var in domain.CreateInvoiceInput
	if err := decodeJSON(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "bad_request", Message: err.Error()})
		return
	}
	inv, err := h.store.CreateInvoiceFromSalesOrder(in)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, inv)
}

func (h *Handler) payInvoice(w http.ResponseWriter, r *http.Request) {
	var in domain.PayInvoiceInput
	if r.ContentLength > 0 {
		if err := decodeJSON(r, &in); err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: "bad_request", Message: err.Error()})
			return
		}
	}
	inv, err := h.store.PayInvoice(r.PathValue("id"), in)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, inv)
}

// --- Purchase Orders ---

func (h *Handler) listPurchaseOrders(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	items := h.store.ListPurchaseOrders(domain.PurchaseOrderFilter{
		VendorID: q.Get("vendorId"),
		Status:   q.Get("status"),
		Plant:    q.Get("plant"),
	})
	writeJSON(w, http.StatusOK, listResponse[domain.PurchaseOrder]{Items: items, Count: len(items)})
}

func (h *Handler) getPurchaseOrder(w http.ResponseWriter, r *http.Request) {
	po, err := h.store.GetPurchaseOrder(r.PathValue("id"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, po)
}

func (h *Handler) createPurchaseOrder(w http.ResponseWriter, r *http.Request) {
	var in domain.CreatePurchaseOrderInput
	if err := decodeJSON(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "bad_request", Message: err.Error()})
		return
	}
	po, err := h.store.CreatePurchaseOrder(in)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, po)
}

func (h *Handler) receivePurchaseOrder(w http.ResponseWriter, r *http.Request) {
	po, err := h.store.ReceivePurchaseOrder(r.PathValue("id"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, po)
}
