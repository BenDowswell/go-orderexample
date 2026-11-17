package main

import (
	"flag"

	"github.com/BenDowswell/go-orderexample/order"
	"github.com/BenDowswell/go-orderexample/server"
)

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	flag.Parse()
	repo := order.NewMemoryOrderRepository()
	svc := order.NewOrderService(repo)
	server.StartServer(*addr, svc)
}
