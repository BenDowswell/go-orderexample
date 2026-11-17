package server

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/BenDowswell/go-orderexample/order"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestHome(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	home(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	if got, want := res.Body.String(), "Hello from server"; got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}

func TestCreateOrder(t *testing.T) {
	svc := order.NewOrderService(order.NewMemoryOrderRepository())
	handler := routes(svc, testLogger())
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"Customer":"Ben","Amount":42.5}`))
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body: %s", res.Code, http.StatusCreated, res.Body.String())
	}
	if got := res.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}
	var body struct {
		ID      int    `json:"id"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.ID != 1 {
		t.Errorf("id = %d, want 1", body.ID)
	}
	if body.Message != "Order created successfully" {
		t.Errorf("message = %q, want success message", body.Message)
	}
}

func TestCreateOrderRejectsInvalidJSON(t *testing.T) {
	handler := routes(nil, testLogger())
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"Customer":`))
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusBadRequest)
	}
}
func TestCreateOrderRejectsUnknownFields(t *testing.T) {
	handler := routes(nil, testLogger())
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"Customer":"Ben","Amount":42.5,"unexpected":true}`))
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusBadRequest)
	}
}
func TestCreateOrderRejectsInvalidOrder(t *testing.T) {
	svc := order.NewOrderService(order.NewMemoryOrderRepository())
	handler := routes(svc, testLogger())
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"Customer":" ","Amount":42.5}`))
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusBadRequest)
	}
}

type failingOrderRepository struct{}

func (failingOrderRepository) Save(order.Order) (int, error) {
	return 0, errors.New("database unavailable")
}

func (failingOrderRepository) GetAll() ([]order.Order, error) {
	return nil, errors.New("database unavailable")
}
func TestCreateOrderReturnsInternalServerErrorWhenRepositoryFails(t *testing.T) {
	svc := order.NewOrderService(failingOrderRepository{})
	handler := routes(svc, testLogger())
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"Customer":"Ben","Amount":42.5}`))
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d; body: %s", res.Code, http.StatusInternalServerError, res.Body.String())
	}
}

func TestGetOrders(t *testing.T) {
	svc := order.NewOrderService(order.NewMemoryOrderRepository())
	svc.CreateOrder(order.Order{Customer: "Ben", Amount: 42.5})
	svc.CreateOrder(order.Order{Customer: "Alice", Amount: 10})
	handler := routes(svc, testLogger())
	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", res.Code, http.StatusOK, res.Body.String())
	}
	if got := res.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}
	var orders []order.Order
	if err := json.Unmarshal(res.Body.Bytes(), &orders); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	want := []order.Order{
		{ID: 1, Customer: "Ben", Amount: 42.5},
		{ID: 2, Customer: "Alice", Amount: 10},
	}
	if len(orders) != len(want) {
		t.Fatalf("got %d orders, want %d", len(orders), len(want))
	}
	for i := range want {
		if orders[i] != want[i] {
			t.Errorf("orders[%d] = %+v, want %+v", i, orders[i], want[i])
		}
	}
}

func TestGetOrdersEmptyReturnsEmptyArray(t *testing.T) {
	svc := order.NewOrderService(order.NewMemoryOrderRepository())
	handler := routes(svc, testLogger())
	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	if got := strings.TrimSpace(res.Body.String()); got != "[]" {
		t.Fatalf("body = %q, want []", got)
	}
}

func TestGetOrdersReturnsInternalServerErrorWhenRepositoryFails(t *testing.T) {
	svc := order.NewOrderService(failingOrderRepository{})
	handler := routes(svc, testLogger())
	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusInternalServerError)
	}
}
