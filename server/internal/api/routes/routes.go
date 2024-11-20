package routes

import (
	"github.com/ambravo/a0-telegram-bot/internal/api/handlers"
	"github.com/ambravo/a0-telegram-bot/internal/api/middleware"
	"github.com/ambravo/a0-telegram-bot/internal/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func SetupRoutes(r *gin.Engine, cfg *config.Config, logger *zap.Logger) {
	// Middleware to set Logger
	r.Use(middleware.RequestLogger(logger))

	// Container health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Bot routes group
	bot := r.Group("/bot")
	{
		// Telegram updates webhook
		bot.POST("/updates", middleware.ValidateTelegramSecret(cfg.DefaultSecretToken),
			handlers.HandleTelegramUpdates(cfg, logger))

		// Auth form routes
		bot.GET("/auth-form", handlers.RenderAuthForm(cfg, logger))
		bot.POST("/auth-form", handlers.ProcessAuthForm(cfg, logger))
	}

	// Auth0 routes group
	auth0 := r.Group("/auth0")
	{
		// OTP webhook
		auth0.POST("/OTPs", middleware.ValidateHMACToken(cfg.HMACSecret),
			handlers.HandleOTPWebhook(cfg, logger))
	}
}
