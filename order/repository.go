package order

type OrderRepository interface {
	Save(order Order) (int, error)
	GetAll() ([]Order, error)
}
