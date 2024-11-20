// internal/config/config.go

package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config holds all configuration for the application
type Config struct {
	// Server settings
	BotPort int    `json:"bot_port"`
	BaseURL string `json:"base_url"`

	// Security tokens
	TelegramToken      string `json:"telegram_token"`
	DefaultSecretToken string `json:"default_secret_token"`
	HMACSecret         string `json:"hmac_secret"`

	// Auth0 settings
	Auth0DemoPlatformApiURL string `json:"auth0_api_url"`

	// Environment
	Environment string `json:"environment"`
}

// LoadConfig loads the configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// Try to load from config.env if .env doesn't exist
		if err := godotenv.Load("config.env"); err != nil {
			fmt.Printf("Warning: No .env file found\n")
		}
	}

	// Initialize logger based on environment
	var logger *zap.Logger
	env := os.Getenv("GIN_MODE")
	if env == "release" {
		logger, _ = zap.NewProduction()
	} else {
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		logger, _ = config.Build()
	}
	defer logger.Sync()

	cfg := &Config{
		// Default values
		BotPort:            8080,
		DefaultSecretToken: "a-very-long-default-secret-string",
		HMACSecret:         "a-very-long-default-secret-string",
		Environment:        env,
	}

	// Load BOT_PORT with default fallback
	if portStr := os.Getenv("BOT_PORT"); portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			logger.Error("Invalid BOT_PORT value", zap.Error(err))
			return nil, fmt.Errorf("invalid BOT_PORT value: %v", err)
		}
		cfg.BotPort = port
	}

	// Required configurations
	cfg.TelegramToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	if cfg.TelegramToken == "" {
		logger.Error("TELEGRAM_BOT_TOKEN is required")
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}

	// Optional configurations with defaults
	if token := os.Getenv("DEFAULT_SECRET_TOKEN"); token != "" {
		cfg.DefaultSecretToken = token
	}

	if secret := os.Getenv("HMAC_DEFAULT_SECRET"); secret != "" {
		cfg.HMACSecret = secret
	}

	// Base URL for the application
	cfg.BaseURL = fmt.Sprintf("http://localhost:%d", cfg.BotPort)
	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		cfg.BaseURL = baseURL
	}

	logger.Info("Configuration loaded successfully",
		zap.Int("port", cfg.BotPort),
		zap.String("environment", cfg.Environment),
		zap.String("base_url", cfg.BaseURL))

	return cfg, nil
}
