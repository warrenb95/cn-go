package main

import (
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/warrenb95/cn-go/internal/store"
)

func putHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		http.Error(w, "missing key ", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	defer log.Println(r.Body.Close())
	if err != nil {
		http.Error(w, "failed to read body", http.StatusInternalServerError)
		return
	}

	err = store.Put(key, string(body))
	if err != nil {
		http.Error(w, "failed to save key: value", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		http.Error(w, "missing key", http.StatusBadRequest)
		return
	}

	value, err := store.Get(key)
	if err != nil {
		if errors.Is(err, store.ErrKeyNotFound) {
			http.Error(w, "key value not found", http.StatusNotFound)
			return
		}
		http.Error(w, "can't get key value", http.StatusInternalServerError)
		return
	}

	_, err = w.Write([]byte(value))
	if err != nil {
		http.Error(w, "failed to write key value", http.StatusInternalServerError)
		return
	}
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		http.Error(w, "missing key", http.StatusBadRequest)
		return
	}

	err := store.Delete(key)
	if err != nil {
		http.Error(w, "failed to delete key value", http.StatusInternalServerError)
		return
	}
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("PUT /v1/{key}", putHandler)
	mux.HandleFunc("GET /v1/{key}", getHandler)
	mux.HandleFunc("DELETE /v1/{key}", deleteHandler)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
