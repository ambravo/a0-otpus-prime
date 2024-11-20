package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ambravo/a0-telegram-bot/internal/config"
	"github.com/ambravo/a0-telegram-bot/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestHandleOTPWebhook(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{
		TelegramToken: "test-token",
		HMACSecret:    "test-secret",
	}

	tests := []struct {
		name           string
		body           map[string]interface{}
		domain         string
		expectedStatus int
		bearerToken    string
	}{
		{
			name: "Valid OTP event",
			body: map[string]interface{}{
				"tenant_id": "test-tenant",
				"domain":    "test.auth0.com",
				"code":      "123456",
				"message":   "Your verification code is: 123456",
				"raw_event": map[string]interface{}{
					"chat_id": 123456789,
				},
			},
			domain:         "test.auth0.com",
			expectedStatus: 200,
			bearerToken:    utils.GenerateAuth0DomainToken("test.auth0.com", "test-secret"),
		},
		{
			name: "Invalid bearer token",
			body: map[string]interface{}{
				"tenant_id": "test-tenant",
				"domain":    "test.auth0.com",
			},
			domain:         "test.auth0.com",
			expectedStatus: 401,
			bearerToken:    "invalid-token",
		},
		{
			name: "Missing chat_id",
			body: map[string]interface{}{
				"tenant_id": "test-tenant",
				"domain":    "test.auth0.com",
				"raw_event": map[string]interface{}{},
			},
			domain:         "test.auth0.com",
			expectedStatus: 400,
			bearerToken:    utils.GenerateAuth0DomainToken("test.auth0.com", "test-secret"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Create request with body
			bodyBytes, _ := json.Marshal(tt.body)
			c.Request, _ = http.NewRequest(
				http.MethodPost,
				"/auth0/OTPs",
				bytes.NewBuffer(bodyBytes),
			)

			// Set headers
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("Authorization", "Bearer "+tt.bearerToken)
			c.Request.Header.Set("X-Auth0-Domain", tt.domain)

			// Execute handler
			HandleOTPWebhook(cfg, logger)(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
