package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ambravo/a0-telegram-bot/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"strings"
	"time"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func RequestLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		// Only log detailed information if logger is in debug mode
		if ce := logger.Check(zap.DebugLevel, "request details"); ce != nil && c.Request.Method == "POST" {
			// Capture request body
			var bodyBytes []byte
			if c.Request.Body != nil {
				bodyBytes, _ = io.ReadAll(c.Request.Body)
				// Restore the request body for further processing
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}

			// Collect headers
			headers := make(map[string][]string)
			for k, v := range c.Request.Header {
				headers[k] = v
			}

			// Parse request body
			var bodyJSON interface{}
			if len(bodyBytes) > 0 {
				if err := json.Unmarshal(bodyBytes, &bodyJSON); err != nil {
					bodyJSON = string(bodyBytes)
				}
			}

			// Log request details
			logger.Debug("incoming request",
				zap.String("path", path),
				zap.String("method", c.Request.Method),
				zap.Any("headers", headers),
				zap.Any("query_params", c.Request.URL.Query()),
				zap.Any("body", bodyJSON),
			)
		}

		// Capture the response
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		// Log the response
		duration := time.Since(start)
		status := c.Writer.Status()

		if ce := logger.Check(zap.DebugLevel, "response details"); ce != nil {
			// Parse response body
			var responseJSON interface{}
			responseBody := blw.body.String()
			if err := json.Unmarshal(blw.body.Bytes(), &responseJSON); err != nil {
				responseJSON = responseBody
			}

			logger.Debug("outgoing response",
				zap.String("path", path),
				zap.String("method", c.Request.Method),
				zap.Int("status", status),
				zap.Duration("latency", duration),
				zap.Any("response_body", responseJSON),
			)
		} else {
			// Basic info logging for non-debug mode
			logger.Info("request processed",
				zap.String("path", path),
				zap.String("method", c.Request.Method),
				zap.Int("status", status),
				zap.Duration("latency", duration),
			)
		}
	}
}
func ValidateTelegramSecret(expectedToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("x-telegram-bot-api-secret-token")
		if token != expectedToken {
			c.AbortWithStatus(401)
			return
		}
		c.Next()
	}
}
func ValidateHMACToken(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger, _ := zap.NewProduction()
		defer logger.Sync()

		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Error("Missing Authorization header")
			c.AbortWithStatus(401)
			return
		}

		// Check Bearer token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Error("Invalid Authorization header format")
			c.AbortWithStatus(401)
			return
		}

		token := parts[1]

		// Get domain from request
		domain := c.GetHeader("x-auth0-domain")
		if domain == "" {
			logger.Error("Missing x-auth0-domain header")
			c.AbortWithStatus(401)
			return
		}

		// Get domain from request
		chatID := c.GetHeader("x-chat_id")
		if domain == "" {
			logger.Error("Missing x-auth0-domain header")
			c.AbortWithStatus(401)
			return
		}

		// Validate the token
		tokenValue := fmt.Sprintf("%s:%s", domain, chatID)
		expectedToken := utils.GenerateAuth0DomainToken(tokenValue, secret)
		if token != expectedToken {
			logger.Error("Invalid HMAC token")
			logger.Debug("Invalid HMAC token",
				zap.String("expected", expectedToken),
				zap.String("received", token))
			c.AbortWithStatus(401)
			return
		}

		// Store validated domain in context for later use
		c.Set("auth0_domain", domain)
		c.Set("chat_id", chatID)
		c.Next()
	}
}
