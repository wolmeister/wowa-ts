package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

type KeyValueEntry struct {
	Key   []string `json:"key"`
	Value string   `json:"value"`
}

type KeyValueStore struct {
	storePath   string
	entries     []KeyValueEntry
	mu          sync.Mutex
	initialized bool
}

func NewKeyValueStore(storePath string) (*KeyValueStore, error) {
	var kvs = &KeyValueStore{storePath: storePath}

	kvs.mu.Lock()
	defer kvs.mu.Unlock()

	if kvs.initialized {
		return nil, errors.New("the KeyValueStore is already initialized")
	}
	kvs.initialized = true
	kvs.storePath = storePath

	if _, err := os.Stat(storePath); err == nil {
		data, err := os.ReadFile(storePath)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(data, &kvs.entries)
		if err != nil {
			return nil, err
		}
	}

	return kvs, nil
}

func keyEquals(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (kvs *KeyValueStore) Get(key []string) (string, error) {
	kvs.mu.Lock()
	defer kvs.mu.Unlock()

	if !kvs.initialized {
		return "", errors.New("the KeyValueStore is not initialized yet")
	}

	for _, entry := range kvs.entries {
		if keyEquals(entry.Key, key) {
			return entry.Value, nil
		}
	}

	return "", nil
}

func (kvs *KeyValueStore) GetByPrefix(prefix []string) ([]string, error) {
	kvs.mu.Lock()
	defer kvs.mu.Unlock()

	if !kvs.initialized {
		return nil, errors.New("the KeyValueStore is not initialized yet")
	}

	var values []string
	for _, entry := range kvs.entries {
		if keyEquals(entry.Key[:len(prefix)], prefix) {
			values = append(values, entry.Value)
		}
	}
	return values, nil
}

func (kvs *KeyValueStore) Set(key []string, value *string) error {
	kvs.mu.Lock()
	defer kvs.mu.Unlock()

	if !kvs.initialized {
		return errors.New("the KeyValueStore is not initialized yet")
	}

	// Remove existing entry with the same key
	for i := 0; i < len(kvs.entries); i++ {
		if keyEquals(kvs.entries[i].Key, key) {
			kvs.entries = append(kvs.entries[:i], kvs.entries[i+1:]...)
			break
		}
	}

	if value != nil {
		kvs.entries = append(kvs.entries, KeyValueEntry{Key: key, Value: *value})
	}

	// Ensure the directory exists
	dir := filepath.Dir(kvs.storePath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	// Write entries to file
	data, err := json.MarshalIndent(kvs.entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(kvs.storePath, data, os.ModePerm)
}
