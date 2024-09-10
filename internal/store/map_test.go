package store_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warrenb95/cn-go/internal/store"
)

func TestPut(t *testing.T) {
	t.Parallel()
	// Config

	// Test Cases
	tests := map[string]struct {
		key, value  string
		errContains string
	}{
		"pass": {
			key:   "key",
			value: "value",
		},
	}

	// Testing
	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s := store.NewMapStore()
			err := s.Put(test.key, test.value)
			if test.errContains != "" {
				require.Error(t, err, "put error")
				assert.ErrorContains(t, err, test.errContains, "put error")
				return
			}
			require.NoError(t, err, "put error")

			value, err := s.Get(test.key)
			require.NoError(t, err, "getting the value")

			assert.Equal(t, test.value, value, "put value")
		})
	}
}

func TestGet(t *testing.T) {
	t.Parallel()
	// Config

	// Test Cases
	tests := map[string]struct {
		storedKey, storedValue string
		key, expectedValue     string
		errContains            string
	}{
		"pass": {
			storedKey:     "key",
			storedValue:   "value",
			key:           "key",
			expectedValue: "value",
		},
		"not found": {
			key:         "not found",
			errContains: store.ErrKeyNotFound.Error(),
		},
	}

	// Testing
	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			s := store.NewMapStore()

			if test.storedKey != "" {
				err := s.Put(test.storedKey, test.storedValue)
				require.NoError(t, err, "putting stored values")
			}

			got, err := s.Get(test.key)
			if test.errContains != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.errContains)
				return
			}
			require.NoError(t, err)

			assert.NotEmpty(t, got, "get response")
			assert.Equal(t, test.expectedValue, got, "expected value")
		})
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()
	// Config

	// Test Cases
	tests := map[string]struct {
		storedKey, storedValue string
		key                    string
		errContains            string
	}{
		"pass": {
			storedKey:   "delete me",
			storedValue: "value",
			key:         "delete me",
		},
	}

	// Testing
	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s := store.NewMapStore()

			if test.storedKey != "" {
				err := s.Put(test.storedKey, test.storedValue)
				require.NoError(t, err, "putting stored values")
			}

			got, err := s.Get(test.key)
			require.NoError(t, err, "checking value in store before delete")
			require.NotEmpty(t, got, "get response")
			require.Equal(t, test.storedValue, got, "stored value")

			err = s.Delete(test.key)
			if test.errContains != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.errContains)
				return
			}
			require.NoError(t, err)

			_, err = s.Get(test.key)
			require.Error(t, err, "get error after value delete")
			assert.ErrorContains(t, err, store.ErrKeyNotFound.Error(), "get error after deleting value")
		})
	}
}
