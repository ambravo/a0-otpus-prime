package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// GenerateHMAC creates an HMAC signature for the given data using the provided secret
func GenerateHMAC(data, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// ValidateHMAC checks if the provided token matches the expected HMAC signature
func ValidateHMAC(data, token, secret string) bool {
	expectedMAC := GenerateHMAC(data, secret)
	return hmac.Equal([]byte(expectedMAC), []byte(token))
}

// GenerateAuth0DomainToken generates a bearer token for Auth0 domain and Chat Id validation
func GenerateAuth0DomainToken(value, secret string) string {
	return GenerateHMAC(value, secret)
}
