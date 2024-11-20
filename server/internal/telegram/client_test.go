package telegram

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSendMessage(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "/bot123/sendMessage", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		// Parse request body
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		// Verify request body
		assert.Equal(t, float64(123456), body["chat_id"])
		assert.Equal(t, "test message", body["text"])
		assert.Equal(t, "HTML", body["parse_mode"])

		// Send response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok": true,
			"result": map[string]interface{}{
				"message_id": 1,
			},
		})
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewClient("123")
	client.client.SetBaseURL(server.URL)

	// Test sending message
	err := client.SendMessage(123456, "test message")
	assert.NoError(t, err)
}

func TestSendMessageWithKeyboard(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "/bot123/sendMessage", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		// Parse request body
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		// Verify request body
		assert.Equal(t, float64(123456), body["chat_id"])
		assert.Equal(t, "test message", body["text"])
		assert.NotNil(t, body["reply_markup"])

		// Send response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok": true,
			"result": map[string]interface{}{
				"message_id": 1,
			},
		})
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewClient("123")
	client.client.SetBaseURL(server.URL)

	// Test sending message with keyboard
	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{
				{"text": "Button 1", "callback_data": "btn1"},
			},
		},
	}
	err := client.SendMessageWithKeyboard(123456, "test message", keyboard)
	assert.NoError(t, err)
}

func TestSendMessageError(t *testing.T) {
	// Create test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":          false,
			"description": "Bad Request: chat not found",
		})
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewClient("123")
	client.client.SetBaseURL(server.URL)

	// Test sending message
	err := client.SendMessage(123456, "test message")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "telegram API error")
}
