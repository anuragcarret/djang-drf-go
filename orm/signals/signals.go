package signals

import (
	"sync"
)

// Signal represents a lifecycle event
type Signal string

const (
	PostSave   Signal = "post_save"
	PostDelete Signal = "post_delete"
	PreSave    Signal = "pre_save"
	PreDelete  Signal = "pre_delete"
)

// Receiver is a function that handles a signal
type Receiver func(sender interface{}, instance interface{}, kwargs map[string]interface{})

var (
	receivers = make(map[Signal]map[string][]Receiver) // signal -> modelName -> receivers
	mu        sync.RWMutex
)

// Register attaches a receiver to a signal for a specific model
// Use "*" for all models
func Register(signal Signal, modelName string, receiver Receiver) {
	mu.Lock()
	defer mu.Unlock()

	if receivers[signal] == nil {
		receivers[signal] = make(map[string][]Receiver)
	}
	receivers[signal][modelName] = append(receivers[signal][modelName], receiver)
}

// Send dispatches a signal to all registered receivers
func Send(signal Signal, sender interface{}, instance interface{}, kwargs map[string]interface{}) {
	mu.RLock()
	defer mu.RUnlock()

	modelName := "*"
	if instance != nil {
		// Try to determine model name from instance if possible
		// (Simplified for now: user provides modelName if needed or we use reflection)
	}

	// Dispatch to specific model receivers
	// For simplicity, we'll assume kwargs["model"] contains the model name if provided
	if name, ok := kwargs["model"].(string); ok {
		modelName = name
	}

	for _, receiver := range receivers[signal][modelName] {
		receiver(sender, instance, kwargs)
	}

	// Also dispatch to global receivers
	if modelName != "*" {
		for _, receiver := range receivers[signal]["*"] {
			receiver(sender, instance, kwargs)
		}
	}
}
