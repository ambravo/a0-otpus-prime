package telegram

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
	"time"
)

type SendMessageOptions struct {
	ExpireIn time.Duration
}

type SendMessageRequest struct {
	ChatID             int64               `json:"chat_id"`
	Text               string              `json:"text"`
	MessageID          int64               `json:"message_id,omitempty"`
	ParseMode          string              `json:"parse_mode,omitempty"`
	LinkPreviewOptions *LinkPreviewOptions `json:"link_preview_options,omitempty"`
	ReplyMarkup        *ReplyMarkup        `json:"reply_markup,omitempty"`
}

type LinkPreviewOptions struct {
	IsDisabled       bool   `json:"is_disabled,omitempty"`
	Url              string `json:"url,omitempty"`
	PreferSmallMedia bool   `json:"prefer_small_media,omitempty"`
	PreferLargeMedia bool   `json:"prefer_large_media,omitempty"`
	ShowAboveText    bool   `json:"show_above_text,omitempty"`
}

// Define DeleteMessageRequest to structure the request for deleting a message
type DeleteMessageRequest struct {
	ChatID    int64 `json:"chat_id"`
	MessageID int   `json:"message_id"`
}

type ReplyMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard,omitempty"`
}

type InlineKeyboardButton struct {
	Text         string `json:"text"`
	URL          string `json:"url,omitempty"`
	CallbackData string `json:"callback_data,omitempty"`
}

type Client struct {
	token  string
	client *resty.Client
	logger *zap.Logger
}

func NewClient(token string) *Client {
	logger, _ := zap.NewProduction()

	client := resty.New().
		SetBaseURL(fmt.Sprintf("https://api.telegram.org/bot%s", token)).
		SetTimeout(10 * time.Second).
		SetRetryCount(3).
		SetRetryWaitTime(100 * time.Millisecond).
		SetRetryMaxWaitTime(2000 * time.Millisecond)

	return &Client{
		token:  token,
		client: client,
		logger: logger,
	}
}

// postMessage handles both sending and editing a message, sets the expiration time.
func (c *Client) postMessage(endpoint string, req SendMessageRequest, options ...SendMessageOptions) error {
	defaultOptions := SendMessageOptions{
		ExpireIn: 5,
	}
	if len(options) > 0 {
		defaultOptions = options[0]
	}

	resp, err := c.client.R().
		SetBody(req).
		Post(fmt.Sprintf("/%s", endpoint))

	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("telegram API error: %s", string(resp.Body()))
	}

	// Simplified structure to parse the Telegram API response
	type TelegramMessageResponse struct {
		Result struct {
			MessageID int `json:"message_id"`
			Chat      struct {
				ID int64 `json:"id"`
			} `json:"chat"`
		} `json:"result"`
	}

	// Extract MessageID from the response
	var messageResponse TelegramMessageResponse
	if err := json.Unmarshal(resp.Body(), &messageResponse); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Schedule message deletion after the specified expiration time
	if defaultOptions.ExpireIn > 0 {
		time.AfterFunc(defaultOptions.ExpireIn*time.Minute, func() {
			_ = c.DeleteMessage(messageResponse.Result.Chat.ID, messageResponse.Result.MessageID) // Ignoring errors
		})
	}

	return nil
}

// SendMessage sends a message to a chat. The markup parameter is optional.
func (c *Client) SendMessage(chatID int64, text string, markup ...*ReplyMarkup) error {
	req := SendMessageRequest{
		ChatID: chatID,
		Text:   text,
		LinkPreviewOptions: &LinkPreviewOptions{
			IsDisabled: true,
		},
		ParseMode: "HTML",
	}

	// Add markup if provided
	if len(markup) > 0 && markup[0] != nil {
		req.ReplyMarkup = markup[0]
	}

	return c.postMessage("sendMessage", req)
}

// EditMessageText edits a message sent by us. Can be used to remove keyboards.
func (c *Client) EditMessageText(chatID int64, messageID int64, text string, markup ...*ReplyMarkup) error {
	req := SendMessageRequest{
		ChatID:    chatID,
		MessageID: messageID,
		Text:      text,
		LinkPreviewOptions: &LinkPreviewOptions{
			IsDisabled: true,
		},
		ParseMode: "HTML",
	}

	// Add markup if provided
	if len(markup) > 0 && markup[0] != nil {
		req.ReplyMarkup = markup[0]
	}

	return c.postMessage("editMessageText", req)
}

// DeleteMessage deletes a message from a chat
func (c *Client) DeleteMessage(chatID int64, messageID int) error {
	req := DeleteMessageRequest{
		ChatID:    chatID,
		MessageID: messageID,
	}

	resp, err := c.client.R().
		SetBody(req).
		Post("/deleteMessage")

	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("telegram API error: %s", string(resp.Body()))
	}

	c.logger.Debug("Message deleted", zap.Int64("chat_id", chatID), zap.Int("message_id", messageID))

	return nil
}
