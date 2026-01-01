package signals

import (
	"testing"
)

func TestSignals(t *testing.T) {
	t.Run("receives post_save signal", func(t *testing.T) {
		received := false
		receiver := func(sender interface{}, instance interface{}, kwargs map[string]interface{}) {
			received = true
			if kwargs["created"] != true {
				t.Error("expected created=true in kwargs")
			}
		}

		Register(PostSave, "mock_model", receiver)

		Send(PostSave, nil, nil, map[string]interface{}{"model": "mock_model", "created": true})

		if !received {
			t.Error("Signal was not received")
		}
	})

	t.Run("global receiver works", func(t *testing.T) {
		receivedCount := 0
		receiver := func(sender interface{}, instance interface{}, kwargs map[string]interface{}) {
			receivedCount++
		}

		Register(PostSave, "*", receiver)

		Send(PostSave, nil, nil, map[string]interface{}{"model": "any_model"})

		if receivedCount == 0 {
			t.Error("Global signal was not received")
		}
	})
}
