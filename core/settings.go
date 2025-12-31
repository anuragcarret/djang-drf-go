package core

import (
	"encoding/json"
	"os"
	"sync"
)

// Settings handles the global configuration for the framework.
type Settings struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

var globalSettings = &Settings{
	data: make(map[string]interface{}),
}

// LoadFromJSON loads settings from a JSON file.
func LoadFromJSON(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return err
	}

	globalSettings.mu.Lock()
	defer globalSettings.mu.Unlock()
	for k, v := range data {
		globalSettings.data[k] = v
	}
	return nil
}

// Get retrieves a setting by key.
func Get(key string) interface{} {
	globalSettings.mu.RLock()
	defer globalSettings.mu.RUnlock()
	return globalSettings.data[key]
}

// Set manually sets a setting (e.g. for testing).
func Set(key string, value interface{}) {
	globalSettings.mu.Lock()
	defer globalSettings.mu.Unlock()
	globalSettings.data[key] = value
}

// GetString helper for string settings.
func GetString(key string) string {
	val := Get(key)
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}
