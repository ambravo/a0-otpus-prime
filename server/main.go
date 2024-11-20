// cmd/server/main.go

package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ambravo/a0-OTPus-prime/server/internal/api/routes"
	"github.com/ambravo/a0-OTPus-prime/server/internal/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//go:embed internal/assets/templates
var templatesFS embed.FS

func main() {
	// Initialize logger based on environment
	var logger *zap.Logger
	if gin.Mode() == gin.ReleaseMode {
		logger, _ = zap.NewProduction()
	} else {
		logConfig := zap.NewDevelopmentConfig()
		logConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		logger, _ = logConfig.Build()
	}
	defer logger.Sync()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Set Gin mode
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize router
	router := gin.New()
	router.Use(gin.Recovery())

	// Setup routes
	routes.SetupRoutes(router, cfg, logger)

	// Create server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.BotPort),
		Handler: router,
	}

	// Server startup in a goroutine
	go func() {
		logger.Info("Starting server",
			zap.Int("port", cfg.BotPort),
			zap.String("environment", cfg.Environment))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited properly")
}
