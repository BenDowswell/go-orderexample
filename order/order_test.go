package order

import (
	"errors"
	"sync"
	"testing"
)

func TestCreateOrder_Success(t *testing.T) {
	repo := NewMemoryOrderRepository()
	svc := NewOrderService(repo)
	order := Order{Customer: "Alice", Amount: 99.99}

	id, err := svc.CreateOrder(order)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if id != 1 {
		t.Errorf("expected ID 1, got %d", id)
	}

	stored, exists := repo.store[id]
	if !exists {
		t.Fatalf("expected order to be stored in DB")
	}

	if stored.Customer != "Alice" || stored.Amount != 99.99 {
		t.Errorf("stored order details mismatch: %+v", stored)
	}
}

func TestCreateOrder_Validation(t *testing.T) {
	repo := NewMemoryOrderRepository()
	svc := NewOrderService(repo)

	tests := []struct {
		name  string
		order Order
	}{
		{"Empty Customer", Order{Customer: "", Amount: 50.0}},
		{"Whitespace Customer", Order{Customer: "   ", Amount: 50.0}},
		{"Zero Amount", Order{Customer: "Bob", Amount: 0}},
		{"Negative Amount", Order{Customer: "Charlie", Amount: -10.0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := svc.CreateOrder(tt.order)
			if !errors.Is(err, ErrValidation) {
				t.Errorf("expected ErrValidation, got %v", err)
			}
			if id != 0 {
				t.Errorf("expected id 0 on error, got %d", id)
			}
		})
	}
}

func TestCreateOrder_Concurrent(t *testing.T) {
	repo := NewMemoryOrderRepository()
	svc := NewOrderService(repo)
	var wg sync.WaitGroup
	workers := 100

	// Launch multiple goroutines to create orders concurrently
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(amt float64) {
			defer wg.Done()
			_, err := svc.CreateOrder(Order{Customer: "Concurrent Buyer", Amount: amt})
			if err != nil {
				t.Errorf("unexpected error in concurrent create: %v", err)
			}
		}(float64(i + 1))
	}

	wg.Wait()

	// Verify that all orders were safely stored without race conditions
	if len(repo.store) != workers {
		t.Errorf("expected %d items in store, got %d", workers, len(repo.store))
	}

	// Verify idCounter incremented correctly
	if repo.nextID != workers+1 {
		t.Errorf("expected idCounter to be %d, got %d", workers+1, repo.nextID)
	}
}

func TestGetAllOrders(t *testing.T) {
	svc := NewOrderService(NewMemoryOrderRepository())

	orders, err := svc.GetAllOrders()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(orders) != 0 {
		t.Fatalf("expected no orders, got %d", len(orders))
	}

	svc.CreateOrder(Order{Customer: "Alice", Amount: 10})
	svc.CreateOrder(Order{Customer: "Bob", Amount: 20})

	orders, err = svc.GetAllOrders()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []Order{
		{ID: 1, Customer: "Alice", Amount: 10},
		{ID: 2, Customer: "Bob", Amount: 20},
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
