package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/ambravo/a0-telegram-bot/internal/assets"
	"github.com/ambravo/a0-telegram-bot/internal/auth0"
	"github.com/ambravo/a0-telegram-bot/internal/config"
	"github.com/ambravo/a0-telegram-bot/internal/telegram"
	"github.com/ambravo/a0-telegram-bot/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthFormData struct {
	ChatID    string
	Signature string
	AuthType  string
	CSRFToken string
}

type AuthFormRequest struct {
	ChatID       string `json:"chat_id" binding:"required"`
	Signature    string `json:"signature" binding:"required"`
	AuthType     string `json:"auth_type" binding:"required"`
	Domain       string `json:"domain" binding:"required"`
	MessageID    string `json:"message_id" binding:"required"`
	AccessToken  string `json:"access_token"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func RenderAuthForm(cfg *config.Config, logger *zap.Logger) gin.HandlerFunc {
	templatesFS := &assets.Assets

	// Parse template at initialization
	tmpl, err := template.ParseFS(templatesFS, "templates/auth_form.html")
	if err != nil {
		logger.Fatal("Failed to parse template",
			zap.Error(err))
	}

	return func(c *gin.Context) {
		chatID := c.Query("chat_id")
		messageID := c.Query("message_id")
		signature := c.Query("signature")
		authType := c.Query("auth_type")

		// Validate required parameters
		if chatID == "" || signature == "" || authType == "" || messageID == "" {
			logger.Error("Missing required parameters",
				zap.String("chat_id", chatID),
				zap.String("auth_type", authType))
			c.String(http.StatusBadRequest, "Invalid request parameters")
			return
		}

		// Validate signature
		if !utils.ValidateHMAC(chatID, signature, cfg.HMACSecret) {
			logger.Error("Invalid signature for auth form",
				zap.String("chat_id", chatID),
				zap.String("signature", signature))
			c.String(http.StatusUnauthorized, "Invalid request signature")
			return
		}

		// Remove keyboard from chat
		telegramClient := telegram.NewClient(cfg.TelegramToken)
		chatIDInt, _ := strconv.ParseInt(chatID, 10, 64)
		messageIDInt, _ := strconv.ParseInt(messageID, 10, 64)
		err = telegramClient.EditMessageText(chatIDInt, messageIDInt, "Continuing in the browser...")
		if err != nil {
			logger.Info("Failed to send message",
				zap.Error(err),
				zap.Int64("chat_id", chatIDInt),
				zap.Int64("message_id", messageIDInt),
			)
		}

		// Generate CSRF token
		csrfToken := utils.GenerateRandomString(32)
		c.SetCookie("csrf_token", csrfToken, 3600, "/", "", true, true)

		// Prepare template data with proper JSON encoding
		data := struct {
			ChatID    template.JS
			MessageID template.JS
			Signature template.JS
			AuthType  template.JS
			CSRFToken template.JS
		}{
			ChatID:    template.JS(chatID),
			MessageID: template.JS(messageID),
			Signature: template.JS(signature),
			AuthType:  template.JS(authType),
			CSRFToken: template.JS(csrfToken),
		}

		// Set security headers
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdnjs.cloudflare.com https://unpkg.com;"+
				"style-src 'self' 'unsafe-inline' https://cdnjs.cloudflare.com;")

		// Render template
		if err := tmpl.Execute(c.Writer, data); err != nil {
			logger.Error("Failed to render template",
				zap.Error(err))
			c.String(http.StatusInternalServerError, "Internal server error")
			return
		}
	}
}
func ProcessAuthForm(cfg *config.Config, logger *zap.Logger) gin.HandlerFunc {
	auth0Client := auth0.NewAuth0Client()
	telegramClient := telegram.NewClient(cfg.TelegramToken)

	return func(c *gin.Context) {
		// Validate CSRF token
		csrfToken, err := c.Cookie("csrf_token")
		if err != nil || csrfToken != c.GetHeader("X-CSRF-Token") {
			logger.Error("CSRF token validation failed", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid CSRF token"})
			return
		}

		var req AuthFormRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Error("Failed to parse auth form request", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		// Validate signature
		if !utils.ValidateHMAC(req.ChatID, req.Signature, cfg.HMACSecret) {
			logger.Error("Invalid signature for auth form submission",
				zap.String("chat_id", req.ChatID))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid request signature"})
			return
		}

		// Validate domain format
		if !utils.IsValidDomain(req.Domain) {
			logger.Error("Invalid domain format", zap.String("domain", req.Domain))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain format"})
			return
		}

		chatIDInt, err := strconv.ParseInt(req.ChatID, 10, 64)
		if err != nil {
			logger.Error("Invalid chat ID format", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat ID format"})
			return
		}

		var accessToken string
		var tokenResponse *auth0.TokenResponse

		// Handle different authentication types
		switch req.AuthType {
		case "tenant_personal":
			deviceCode, err := auth0Client.InitiateDeviceFlow(req.Domain)
			if err != nil {
				logger.Error("Failed to initiate device flow",
					zap.Error(err),
					zap.String("domain", req.Domain))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate device flow"})
				return
			}

			// Send device code info via Telegram
			message := `🔐 To complete the authentication, please use this code:

Code: ` + deviceCode.UserCode + `

Visit: ` + deviceCode.VerificationURIComplete + `

This code will expire in ` + strconv.Itoa(deviceCode.ExpiresIn/60) + ` minutes.`

			err = telegramClient.SendMessage(chatIDInt, message)
			if err != nil {
				logger.Error("Failed to send device code info",
					zap.Error(err),
					zap.Int64("chat_id", chatIDInt))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send authentication information"})
				return
			}

			// Start polling for token
			go pollForDeviceToken(auth0Client, cfg, logger, deviceCode, chatIDInt, req.Domain)

			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "Authentication process initiated. Please check your Telegram chat for instructions.",
			})
			return

		case "auth_ephemeral":
			if req.AccessToken == "" {
				logger.Error("Missing access token for ephemeral auth")
				c.JSON(http.StatusBadRequest, gin.H{"error": "Access token is required"})
				return
			}
			accessToken = req.AccessToken

		case "auth_client_credentials":
			if req.ClientID == "" || req.ClientSecret == "" {
				logger.Error("Missing client credentials")
				c.JSON(http.StatusBadRequest, gin.H{"error": "Client ID and secret are required"})
				return
			}

			tokenResponse, err = auth0Client.GetClientCredentialsToken(req.Domain, req.ClientID, req.ClientSecret)
			if err != nil {
				logger.Error("Failed to get client credentials token",
					zap.Error(err),
					zap.String("domain", req.Domain))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get access token"})
				return
			}
			accessToken = tokenResponse.AccessToken

		default:
			logger.Error("Invalid auth type", zap.String("type", req.AuthType))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid authentication type"})
			return
		}

		// Create or update Auth0 Action
		err = auth0Client.EnablePhoneExtensibility(req.Domain, accessToken, chatIDInt, cfg)
		if err != nil {
			logger.Error("Failed to setup Auth0 action",
				zap.Error(err),
				zap.String("domain", req.Domain),
				zap.Int64("chat_id", chatIDInt))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to setup Auth0 action"})
			return
		}

		// Send success message via Telegram
		message := fmt.Sprintf(
			"✅ Configuration completed successfully!\n\n"+
				"Domain: %s\n"+
				"Action: telegram-otp-action\n\n"+
				"You will now receive OTP codes in this chat.",
			req.Domain,
		)

		messageIDInt, _ := strconv.ParseInt(req.MessageID, 10, 64)
		err = telegramClient.EditMessageText(chatIDInt, messageIDInt, message)
		if err != nil {
			logger.Error("Failed to send success message",
				zap.Error(err),
				zap.Int64("chat_id", chatIDInt))
			// Don't return error as the main setup was successful
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Authentication and setup completed successfully",
		})
	}
}

func pollForDeviceToken(
	client *auth0.Auth0Client,
	cfg *config.Config,
	logger *zap.Logger,
	deviceCode *auth0.DeviceCodeResponse,
	chatID int64,
	domain string,
) {
	telegramClient := telegram.NewClient(cfg.TelegramToken)
	interval := time.Duration(deviceCode.Interval) * time.Second
	expiry := time.Now().Add(time.Duration(deviceCode.ExpiresIn) * time.Second)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if time.Now().After(expiry) {
				message := "❌ Authentication process timed out. Please try again."
				_ = telegramClient.SendMessage(chatID, message)
				return
			}

			token, err := client.PollDeviceToken(domain, deviceCode.DeviceCode)
			if err != nil {
				// Check if it's just a pending authorization
				if err == auth0.ErrAuthorizationPending {
					continue
				}
				logger.Error("Failed to poll for device token",
					zap.Error(err),
					zap.String("domain", domain))
				message := "❌ Authentication failed. Please try again."
				_ = telegramClient.SendMessage(chatID, message)
				return
			}

			// Successfully got token, create action
			err = client.EnablePhoneExtensibility(domain, token.AccessToken, chatID, cfg)
			if err != nil {
				logger.Error("Failed to setup Auth0 action after device flow",
					zap.Error(err),
					zap.String("domain", domain))
				message := "❌ Failed to setup Auth0 action. Please try again."
				_ = telegramClient.SendMessage(chatID, message)
				return
			}

			// Send success message
			message := fmt.Sprintf(
				"✅ Configuration completed successfully!\n\n"+
					"Domain: %s\n"+
					"Action: telegram-otp-action-%d\n\n"+
					"You will now receive OTP codes in this chat.",
				domain, chatID,
			)
			_ = telegramClient.SendMessage(chatID, message)
			return
		}
	}
}
