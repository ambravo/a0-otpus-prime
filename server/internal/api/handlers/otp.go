package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/ambravo/a0-telegram-bot/internal/config"
	"github.com/ambravo/a0-telegram-bot/internal/telegram"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"strconv"
)

// OTPEvent represents the incoming OTP event from Auth0
type OTPEvent struct {
	TenantID    string                 `json:"tenant_id"`
	Domain      string                 `json:"domain,omitempty"`
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`
	PhoneNumber string                 `json:"phone_number"`
	RawEvent    map[string]interface{} `json:"raw_event"`
}

// TelegramMessage represents the structure for sending messages via Telegram API
type TelegramMessageRequest struct {
	ChatID              int64  `json:"chat_id"`
	Text                string `json:"text"`
	ParseMode           string `json:"parse_mode,omitempty"`
	DisableNotification bool   `json:"disable_notification,omitempty"`
}

func HandleOTPWebhook(cfg *config.Config, logger *zap.Logger) gin.HandlerFunc {
	telegramClient := telegram.NewClient(cfg.TelegramToken)

	return func(c *gin.Context) {
		var event OTPEvent
		if err := c.ShouldBindJSON(&event); err != nil {
			logger.Error("Failed to parse OTP event",
				zap.Error(err))
			c.JSON(400, gin.H{"error": "Invalid event format"})
			return
		}

		// Validate domain matches the one from auth header
		var exists bool
		domain, exists := c.Get("auth0_domain")
		if !exists {
			logger.Error("Domain mismatch",
				zap.String("header_domain", domain.(string)),
				zap.String("event_domain", event.Domain))
			c.JSON(400, gin.H{"error": "Domain mismatch"})
			return
		}

		chatIDStr, exists := c.Get("chat_id")
		if !exists {
			logger.Error("Chat ID not found")
			c.JSON(400, gin.H{"error": "Chat ID not found"})
			return
		}
		chatID, _ := strconv.ParseInt(chatIDStr.(string), 10, 64)

		// Prepare Telegram message
		message := formatTelegramMessage(event)

		// Send to Telegram
		err := telegramClient.SendMessage(chatID, message)

		if err != nil {
			logger.Error("Failed to send Telegram message",
				zap.Error(err),
				zap.String("tenant_id", event.TenantID))
			c.JSON(500, gin.H{"error": "Failed to send message"})
			return
		}

		logger.Info("OTP message sent successfully",
			zap.String("tenant_id", event.TenantID),
			zap.Int64("chat_id", chatID))

		c.JSON(200, gin.H{"status": "ok"})
	}
}

// formatTelegramMessage creates a formatted message for Telegram
func formatTelegramMessage(event OTPEvent) string {
	jsonData, _ := json.MarshalIndent(event, "", "  ")
	return fmt.Sprintf(
		""+
			"Domain: <code>%s</code>\n\n"+
			"Recipent: <code>%s</code>\n"+
			"Code: <b><code>%s</code></b>\n"+
			"Message: <code>%s</code>\n\n"+
			"<blockquote expandable>\n<b>All details:</b>\n\n👇👇👇\n\n%s</blockquote>",
		event.Domain,
		event.PhoneNumber,
		event.Code,
		event.Message,
		jsonData,
	)
}
