package utils

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// InitLogger initializes a new logger with the given log level
func InitLogger(level string) (*zap.Logger, error) {
	var zapLevel zapcore.Level
	err := zapLevel.UnmarshalText([]byte(level))
	if err != nil {
		zapLevel = zapcore.InfoLevel // Default to INFO if invalid level
	}

	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapLevel),
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	// Customize time format
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	return config.Build()
}

// CreateLoggerWithContext creates a logger with additional context fields
func CreateLoggerWithContext(baseLogger *zap.Logger, fields ...zapcore.Field) *zap.Logger {
	return baseLogger.With(fields...)
}
