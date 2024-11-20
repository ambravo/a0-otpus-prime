package utils

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateRandomString generates a random string of specified length
func GenerateRandomString(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		panic(err) // Handle this according to your needs
	}
	return base64.URLEncoding.EncodeToString(b)[:length]
}

// IsValidDomain checks if a domain string is valid
func IsValidDomain(domain string) bool {
	// Add your domain validation logic here
	// This is a simple check, you might want to add more validation
	if domain == "" {
		return false
	}
	if len(domain) < 4 { // minimum: a.com
		return false
	}
	if len(domain) > 255 {
		return false
	}
	return true
}
