package store

import (
	"errors"
	"fmt"
	"sync"
)

type MapStore struct {
	sync.RWMutex
	m map[string]string
}

var (
	ErrKeyNotFound = errors.New("key not found")
)

func NewMapStore() *MapStore {
	return &MapStore{
		m: make(map[string]string),
	}
}

func (s *MapStore) Put(key, value string) error {
	s.Lock()
	defer s.Unlock()
	s.m[key] = value

	return nil
}

func (s *MapStore) Get(key string) (string, error) {
	s.RLock()
	defer s.RUnlock()
	v, ok := s.m[key]
	if !ok {
		return "", fmt.Errorf("getting from store: %w", ErrKeyNotFound)
	}

	return v, nil
}

func (s *MapStore) Delete(key string) error {
	s.Lock()
	defer s.Unlock()
	delete(s.m, key)

	return nil
}
