package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHMACFunctions(t *testing.T) {
	tests := []struct {
		name   string
		data   string
		secret string
	}{
		{
			name:   "Basic HMAC generation",
			data:   "test.auth0.com",
			secret: "test-secret",
		},
		{
			name:   "Empty data",
			data:   "",
			secret: "test-secret",
		},
		{
			name:   "Empty secret",
			data:   "test.auth0.com",
			secret: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate HMAC
			token := GenerateHMAC(tt.data, tt.secret)

			// Validate generated HMAC
			assert.True(t, ValidateHMAC(tt.data, token, tt.secret))

			// Validate with wrong data
			assert.False(t, ValidateHMAC(tt.data+"wrong", token, tt.secret))

			// Validate with wrong secret
			assert.False(t, ValidateHMAC(tt.data, token, tt.secret+"wrong"))
		})
	}
}

func TestGenerateAuth0DomainToken(t *testing.T) {
	tests := []struct {
		name     string
		domain   string
		secret   string
		expected string
	}{
		{
			name:   "Clean domain",
			domain: "test.auth0.com",
			secret: "test-secret",
		},
		{
			name:   "Domain with https",
			domain: "https://test.auth0.com",
			secret: "test-secret",
		},
		{
			name:   "Domain with trailing slash",
			domain: "test.auth0.com/",
			secret: "test-secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := GenerateAuth0DomainToken(tt.domain, tt.secret)
			cleanDomain := "test.auth0.com" // All test cases should resolve to this
			expectedToken := GenerateHMAC(cleanDomain, tt.secret)
			assert.Equal(t, expectedToken, token)
		})
	}
}
