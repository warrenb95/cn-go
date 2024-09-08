package store

import (
	"errors"
	"fmt"
	"sync"
)

type mapStore struct {
	sync.RWMutex
	m map[string]string
}

var ms = mapStore{
	m: make(map[string]string),
}

var (
	ErrKeyNotFound = errors.New("key not found")
)

func Put(key, value string) error {
	ms.Lock()
	defer ms.Unlock()
	ms.m[key] = value

	return nil
}

func Get(key string) (string, error) {
	ms.RLock()
	defer ms.RUnlock()
	v, ok := ms.m[key]
	if !ok {
		return "", fmt.Errorf("getting from store: %w", ErrKeyNotFound)
	}

	return v, nil
}

func Delete(key string) error {
	ms.Lock()
	defer ms.Unlock()
	delete(ms.m, key)

	return nil
}
