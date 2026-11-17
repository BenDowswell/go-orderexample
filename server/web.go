package server

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/BenDowswell/go-orderexample/order"
)

func StartServer(addr string, svc *order.OrderService) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	srv := &http.Server{
		Addr:    addr,
		Handler: routes(svc, logger),
	}
	logger.Info("Starting server", "addr", addr)
	err := srv.ListenAndServe()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func routes(svc *order.OrderService, logger *slog.Logger) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", home)
	mux.HandleFunc("POST /orders", handleCreateOrder(svc, logger))
	mux.HandleFunc("GET /orders", handleGetOrders(svc, logger))
	return mux
}

func handleGetOrders(svc *order.OrderService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orders, err := svc.GetAllOrders()
		if err != nil {
			logger.Error("Failed to get orders", "error", err)
			http.Error(w, "Failed to get orders", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(orders)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from server"))
}

// handleCreateOrder is a closure that captures svc and logger

func handleCreateOrder(svc *order.OrderService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ord order.Order
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&ord); err != nil {
			logger.Warn("Failed to decode order JSON", "error", err)
			http.Error(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
			return
		}

		id, err := svc.CreateOrder(ord)

		if err != nil {
			if errors.Is(err, order.ErrValidation) {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			logger.Error("Failed to create order", "error", err)
			http.Error(w, "Failed to create order", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{
			"id":      id,
			"message": "Order created successfully",
		})
	}
}
