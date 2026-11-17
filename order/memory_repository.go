package order

import (
	"sync"
)

type MemoryOrderRepository struct {
	mu     sync.Mutex
	store  map[int]Order
	nextID int
}

func NewMemoryOrderRepository() *MemoryOrderRepository {
	return &MemoryOrderRepository{
		store:  make(map[int]Order),
		nextID: 1,
	}
}

func (r *MemoryOrderRepository) Save(order Order) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := r.nextID

	order.ID = id
	r.store[id] = order
	r.nextID++

	return id, nil
}

// GetAll returns all stored orders, ordered by ID (creation order).
func (r *MemoryOrderRepository) GetAll() ([]Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	orders := make([]Order, 0, len(r.store))
	for id := 1; id < r.nextID; id++ {
		if o, ok := r.store[id]; ok {
			orders = append(orders, o)
		}
	}
	return orders, nil
}
