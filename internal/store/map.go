package store

import (
	"errors"
	"fmt"
)

var store = make(map[string]string)

var (
	ErrKeyNotFound = errors.New("key not found")
)

func Put(key, value string) error {
	store[key] = value

	return nil
}

func Get(key string) (string, error) {
	v, ok := store[key]
	if !ok {
		return "", fmt.Errorf("getting from store: %w", ErrKeyNotFound)
	}

	return v, nil
}

func Delete(key string) error {
	delete(store, key)

	return nil
}
