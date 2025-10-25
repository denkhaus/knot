package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func init() {
	// Initialize with no-op logger by default
	// Will be configured later via SetLogLevel when CLI flags are parsed
	Log = zap.NewNop()
}

// GetLogger returns the global logger instance
func GetLogger() *zap.Logger {
	return Log
}

// SetLogLevel configures the global logger with the specified log level
func SetLogLevel(logLevel string) {
	var err error

	switch logLevel {
	case "debug":
		// Development logger for debug mode
		Log, err = zap.NewDevelopment()
		if err != nil {
			Log = zap.NewNop()
		}
	case "info", "warn", "error":
		// For CLI usage, we want human-readable console output
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.TimeKey = ""     // Remove timestamp for cleaner CLI output
		config.EncoderConfig.CallerKey = ""   // Remove caller info for cleaner CLI output
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Colored level names
		config.OutputPaths = []string{"stderr"} // Send to stderr to not interfere with CLI output

		// Set log level
		switch logLevel {
		case "info":
			config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		case "warn":
			config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
		case "error":
			config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
		}

		Log, err = config.Build()
		if err != nil {
			Log = zap.NewNop() // Fallback to no-op logger
		}
	case "off", "":
		// No logging at all (default)
		Log = zap.NewNop()
	default:
		// Invalid log level, use no-op logger
		Log = zap.NewNop()
	}
}

// Sync flushes any buffered log entries
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}