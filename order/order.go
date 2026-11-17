package order

import (
	"errors"
)

var (
	ErrValidation = errors.New("validation error, invalid data")
)

type Order struct {
	ID       int     `json:"id"`
	Customer string  `json:"customer"`
	Amount   float64 `json:"amount"`
}
