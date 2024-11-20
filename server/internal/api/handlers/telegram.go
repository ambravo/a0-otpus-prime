package handlers

import (
	"fmt"
	"github.com/ambravo/a0-telegram-bot/internal/config"
	"github.com/ambravo/a0-telegram-bot/internal/telegram"
	"github.com/ambravo/a0-telegram-bot/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type TelegramUpdate struct {
	UpdateID      int64                  `json:"update_id"`
	Message       *TelegramMessage       `json:"message"`
	CallbackQuery *TelegramCallbackQuery `json:"callback_query"`
}

type TelegramMessage struct {
	MessageID int64            `json:"message_id"`
	From      *TelegramUser    `json:"from"`
	Chat      *TelegramChat    `json:"chat"`
	Text      string           `json:"text"`
	ReplyTo   *TelegramMessage `json:"reply_to_message"`
}

type TelegramCallbackQuery struct {
	ID      string           `json:"id"`
	From    *TelegramUser    `json:"from"`
	Message *TelegramMessage `json:"message"`
	Data    string           `json:"data"`
}

type TelegramUser struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
}

type TelegramChat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

type TelegramReplyMarkup struct {
	Keyboard        [][]TelegramKeyboardButton `json:"keyboard,omitempty"`
	InlineKeyboard  [][]TelegramInlineButton   `json:"inline_keyboard,omitempty"`
	ResizeKeyboard  bool                       `json:"resize_keyboard"`
	OneTimeKeyboard bool                       `json:"one_time_keyboard"`
}

type TelegramKeyboardButton struct {
	Text string `json:"text"`
}

type TelegramInlineButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
	URL          string `json:"url,omitempty"`
}

func HandleTelegramUpdates(cfg *config.Config, logger *zap.Logger) gin.HandlerFunc {
	telegramClient := telegram.NewClient(cfg.TelegramToken)

	return func(c *gin.Context) {
		var update TelegramUpdate
		if err := c.ShouldBindJSON(&update); err != nil {
			logger.Error("Failed to parse telegram update", zap.Error(err))
			c.JSON(400, gin.H{"error": "Invalid update format"})
			return
		}

		// Return 200 OK immediately
		c.Status(200)

		// Handle updates in a goroutine
		go func() {
			if update.CallbackQuery != nil {
				handleCallbackQuery(update.CallbackQuery, cfg, telegramClient, logger)
				return
			}

			if update.Message != nil {
				handleMessage(update.Message, cfg, telegramClient, logger)
				return
			}
		}()
	}
}

// TODO: Re-enable personal and ephemeral auth
func handleMessage(message *TelegramMessage, cfg *config.Config, client *telegram.Client, logger *zap.Logger) {
	if message.Text == "/start" {
		keyboard := &telegram.ReplyMarkup{
			InlineKeyboard: [][]telegram.InlineKeyboardButton{
				{
					{
						Text:         "Personal Public Tenant",
						CallbackData: "auth_client_credentials", //TODO: "tenant_personal",
					},
				},
				{
					{
						Text:         "Private Demo Tenant",
						CallbackData: "auth_client_credentials", //TODO: "tenant_private",
					},
				},
			},
		}

		err := client.SendMessage(message.Chat.ID, "Which kind of instance do you want to connect?", keyboard)
		if err != nil {
			logger.Error("Failed to send message",
				zap.Error(err),
				zap.Int64("chat_id", message.Chat.ID))
		}
	}
}

func handleCallbackQuery(query *TelegramCallbackQuery, cfg *config.Config, client *telegram.Client, logger *zap.Logger) {
	chatID := query.Message.Chat.ID
	messageID := query.Message.MessageID

	switch query.Data {
	case "tenant_personal":
		signature := utils.GenerateHMAC(fmt.Sprintf("%d", chatID), cfg.HMACSecret)
		authURL := fmt.Sprintf("%s/bot/auth-form?chat_id=%d&signature=%s&auth_type=tenant_personal&messageID=%d",
			cfg.BaseURL, chatID, signature, messageID)

		keyboard := &telegram.ReplyMarkup{
			InlineKeyboard: [][]telegram.InlineKeyboardButton{
				{
					{
						Text: "Complete Authentication",
						URL:  authURL,
					},
				},
			},
		}

		err := client.EditMessageText(chatID, messageID, "Please complete your authentication using the form below:", keyboard)
		if err != nil {
			logger.Error("Failed to send message",
				zap.Error(err),
				zap.Int64("chat_id", chatID))
		}

	case "tenant_private":
		keyboard := &telegram.ReplyMarkup{
			InlineKeyboard: [][]telegram.InlineKeyboardButton{
				{
					{
						Text:         "Ephemeral Access Token",
						CallbackData: "auth_ephemeral",
					},
				},
				{
					{
						Text:         "ClientID and Client Credentials",
						CallbackData: "auth_client_credentials",
					},
				},
			},
		}

		err := client.EditMessageText(chatID, messageID, "How would you like to authenticate?", keyboard)
		if err != nil {
			logger.Error("Failed to send message",
				zap.Error(err),
				zap.Int64("chat_id", chatID))
		}

	case "auth_ephemeral", "auth_client_credentials":
		signature := utils.GenerateHMAC(fmt.Sprintf("%d", chatID), cfg.HMACSecret)
		authURL := fmt.Sprintf("%s/bot/auth-form?chat_id=%d&signature=%s&auth_type=%s&message_id=%d",
			cfg.BaseURL, chatID, signature, query.Data, messageID)

		keyboard := &telegram.ReplyMarkup{
			InlineKeyboard: [][]telegram.InlineKeyboardButton{
				{
					{
						Text: "Complete Authentication",
						URL:  authURL,
					},
				},
			},
		}

		err := client.EditMessageText(chatID, messageID, "Please complete your authentication using the form below:", keyboard)
		if err != nil {
			logger.Error("Failed to send message",
				zap.Error(err),
				zap.Int64("chat_id", chatID))
		}
	}
}
