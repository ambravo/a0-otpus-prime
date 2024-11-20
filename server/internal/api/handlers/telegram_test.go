package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ambravo/a0-telegram-bot/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestHandleTelegramUpdates(t *testing.T) {
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
		expectedStatus int
		secretToken    string
	}{
		{
			name: "Valid start command",
			body: map[string]interface{}{
				"message": map[string]interface{}{
					"message_id": 1,
					"chat": map[string]interface{}{
						"id":   123456789,
						"type": "private",
					},
					"text": "/start",
				},
			},
			expectedStatus: 200,
			secretToken:    "test-token",
		},
		{
			name: "Invalid token",
			body: map[string]interface{}{
				"message": map[string]interface{}{
					"text": "/start",
				},
			},
			expectedStatus: 401,
			secretToken:    "wrong-token",
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
				"/bot/updates",
				bytes.NewBuffer(bodyBytes),
			)

			// Set headers
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("x-telegram-bot-api-secret-token", tt.secretToken)

			// Execute handler
			HandleTelegramUpdates(cfg, logger)(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == 200 {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				// Check if response contains reply markup for /start command
				if tt.body["message"].(map[string]interface{})["text"] == "/start" {
					assert.Contains(t, response, "reply_markup")
				}
			}
		})
	}
}
