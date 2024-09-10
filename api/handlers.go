package api

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/warrenb95/cn-go/internal/store"
)

type Store interface {
	Put(key string, value string) error
	Get(key string) (string, error)
	Delete(key string) error
}

type TransactionLogger interface {
	WritePut(key, value string)
	WriteDelete(key string)
	Err() <-chan error

	ReadEvents() (<-chan store.Event, <-chan error)

	Run()
	Close() error
}

type Handler struct {
	txLogger TransactionLogger
	store    Store
}

func NewHandler(txLogger TransactionLogger, store Store) *Handler {
	return &Handler{
		txLogger: txLogger,
		store:    store,
	}
}

func (h *Handler) Close() error {
	err := h.txLogger.Close()
	if err != nil {
		return fmt.Errorf("closing api handler: %w", err)
	}

	return nil
}

func (h *Handler) Put(w http.ResponseWriter, r *http.Request) {
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

	err = h.store.Put(key, string(body))
	if err != nil {
		http.Error(w, "failed to save key: value", http.StatusInternalServerError)
		return
	}
	h.txLogger.WritePut(key, string(body))

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		http.Error(w, "missing key", http.StatusBadRequest)
		return
	}

	value, err := h.store.Get(key)
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

func (h *Handler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		http.Error(w, "missing key", http.StatusBadRequest)
		return
	}

	err := h.store.Delete(key)
	if err != nil {
		http.Error(w, "failed to delete key value", http.StatusInternalServerError)
		return
	}
	h.txLogger.WriteDelete(key)
}
