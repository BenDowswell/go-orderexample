package order

import (
	"strings"
)

type OrderService struct {
	repo OrderRepository
}

func NewOrderService(repo OrderRepository) *OrderService {
	return &OrderService{
		repo: repo,
	}
}

func validateOrder(order Order) error {
	if strings.TrimSpace(order.Customer) == "" {
		return ErrValidation
	}
	if order.Amount <= 0 {
		return ErrValidation
	}
	return nil
}

func (s *OrderService) CreateOrder(order Order) (int, error) {
	if err := validateOrder(order); err != nil {
		return 0, err
	}
	id, err := s.repo.Save(order)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *OrderService) GetAllOrders() ([]Order, error) {
	return s.repo.GetAll()
}
