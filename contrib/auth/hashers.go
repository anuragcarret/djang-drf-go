package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// MakePassword hashes a raw password using bcrypt
func MakePassword(password string) string {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(hashed)
}

// CheckPassword verifies a raw password against a hash
func CheckPassword(password, hashed string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	return err == nil
}
