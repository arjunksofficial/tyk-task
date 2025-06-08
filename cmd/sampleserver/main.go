package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	// Start orders service on port 8000
	go func() {
		orderRouter := mux.NewRouter()
		orderRouter.HandleFunc("/api/v1/orders/list", ordersHandler).Methods("GET")

		log.Println("Orders service running on :8000")
		if err := http.ListenAndServe(":8000", orderRouter); err != nil {
			log.Fatalf("Orders service failed: %v", err)
		}
	}()

	// Start users service on port 8001
	userRouter := mux.NewRouter()
	userRouter.HandleFunc("/api/v1/users/list", usersHandler).Methods("GET")

	log.Println("Users service running on :8001")
	if err := http.ListenAndServe(":8001", userRouter); err != nil {
		log.Fatalf("Users service failed: %v", err)
	}
}

func ordersHandler(w http.ResponseWriter, r *http.Request) {
	orders := []map[string]interface{}{
		{"id": 1, "item": "Book"},
		{"id": 2, "item": "Laptop"},
		{"id": 3, "item": "Pen"},
		{"id": 4, "item": "Mouse"},
		{"id": 5, "item": "Keyboard"},
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"orders": orders})
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	users := []map[string]interface{}{
		{"id": 1, "name": "Alice"},
		{"id": 2, "name": "Bob"},
		{"id": 3, "name": "Charlie"},
		{"id": 4, "name": "Diana"},
		{"id": 5, "name": "Eve"},
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"users": users})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
