package main

import (
	"log"
	"net/http"

	"github.com/warrenb95/cn-go/api"
	"github.com/warrenb95/cn-go/internal/store"
)

func main() {
	mux := http.NewServeMux()

	mapStore := store.NewMapStore()

	txLogger, err := store.NewFileTransactionLogger("tx_logs.log", mapStore)
	if err != nil {
		log.Fatalf("tx logger: %v", err)
	}

	h := api.NewHandler(txLogger, mapStore)
	defer func() {
		err := h.Close()
		if err != nil {
			log.Printf("closing down handler: %v", err)
		}
	}()

	mux.HandleFunc("PUT /v1/{key}", h.Put)
	mux.HandleFunc("GET /v1/{key}", h.Get)
	mux.HandleFunc("DELETE /v1/{key}", h.DeleteHandler)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
