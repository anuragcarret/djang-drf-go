package auth

import (
	"testing"
)

func TestPasswordHashing(t *testing.T) {
	t.Run("hashes and verifies password", func(t *testing.T) {
		password := "secret123"
		hashed := MakePassword(password)

		if hashed == password {
			t.Error("password was not hashed")
		}

		if !CheckPassword(password, hashed) {
			t.Error("password verification failed")
		}

		if CheckPassword("wrong", hashed) {
			t.Error("password verification should have failed for wrong password")
		}
	})
}

func TestUserAuthentication(t *testing.T) {
	// This will require the Model and Registry to be functional
	// For now, let's test the logic with a mock user
}
