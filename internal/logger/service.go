package logger

import (
	"github.com/samber/do/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger defines the logging service interface.
// This abstracts zap.Logger to enable dependency injection and testing.
type Logger interface {
	// Debug logs a debug message with optional fields
	Debug(msg string, fields ...zap.Field)

	// Info logs an info message with optional fields
	Info(msg string, fields ...zap.Field)

	// Warn logs a warning message with optional fields
	Warn(msg string, fields ...zap.Field)

	// Error logs an error message with optional fields
	Error(msg string, fields ...zap.Field)

	// Sync flushes any buffered log entries
	Sync()

	// With adds structured fields to the logger
	With(fields ...zap.Field) Logger

	// Named creates a named logger with the given name
	Named(name string) Logger

	// ToZap returns the underlying zap.Logger for compatibility
	ToZap() *zap.Logger

	// SetLevel configures the logger with the specified log level
	SetLevel(logLevel string)
}

// serviceImpl is the private implementation of the Logger interface
type serviceImpl struct {
	logger *zap.Logger
}

// NewService creates a new logger service instance.
// This follows the dependency injection pattern from di.md.
func NewService(injector do.Injector) (Logger, error) {
	// For now, create a no-op logger. Will be configured later via SetLevel
	logger := zap.NewNop()

	return &serviceImpl{
		logger: logger,
	}, nil
}

// Debug logs a debug message with optional fields
func (s *serviceImpl) Debug(msg string, fields ...zap.Field) {
	s.logger.Debug(msg, fields...)
}

// Info logs an info message with optional fields
func (s *serviceImpl) Info(msg string, fields ...zap.Field) {
	s.logger.Info(msg, fields...)
}

// Warn logs a warning message with optional fields
func (s *serviceImpl) Warn(msg string, fields ...zap.Field) {
	s.logger.Warn(msg, fields...)
}

// Error logs an error message with optional fields
func (s *serviceImpl) Error(msg string, fields ...zap.Field) {
	s.logger.Error(msg, fields...)
}

// Sync flushes any buffered log entries
func (s *serviceImpl) Sync() {
	_ = s.logger.Sync()
}

// With adds structured fields to the logger
func (s *serviceImpl) With(fields ...zap.Field) Logger {
	return &serviceImpl{
		logger: s.logger.With(fields...),
	}
}

// Named creates a named logger with the given name
func (s *serviceImpl) Named(name string) Logger {
	return &serviceImpl{
		logger: s.logger.Named(name),
	}
}

// ToZap returns the underlying zap.Logger for compatibility
func (s *serviceImpl) ToZap() *zap.Logger {
	return s.logger
}

// SetLevel configures the logger with the specified log level
// This method allows updating the log level after service creation
func (s *serviceImpl) SetLevel(logLevel string) {
	var err error

	switch logLevel {
	case "debug":
		s.logger, err = zap.NewDevelopment()
		if err != nil {
			s.logger = zap.NewNop()
		}
	case "info", "warn", "error":
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.TimeKey = ""
		config.EncoderConfig.CallerKey = ""
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.OutputPaths = []string{"stderr"}

		switch logLevel {
		case "info":
			config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		case "warn":
			config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
		case "error":
			config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
		}

		s.logger, err = config.Build()
		if err != nil {
			s.logger = zap.NewNop()
		}
	case "off", "":
		s.logger = zap.NewNop()
	default:
		s.logger = zap.NewNop()
	}
}

// Error creates a zap.Field for error logging
func Error(err error) zap.Field {
	return zap.Error(err)
}

// String creates a zap.Field for string values
func String(key, value string) zap.Field {
	return zap.String(key, value)
}

// Int creates a zap.Field for integer values
func Int(key string, value int) zap.Field {
	return zap.Int(key, value)
}

// Any creates a zap.Field for any value
func Any(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}
