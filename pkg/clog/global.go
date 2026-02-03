package clog

import (
	"context"
	"sync"
)

var (
	globalLogger Logger
	globalMutex  sync.RWMutex
)

// InitGlobalLogger initializes the global logger instance
func InitGlobalLogger(config *Config) error {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	logger, err := NewLogger(config)
	if err != nil {
		return err
	}

	globalLogger = logger
	return nil
}

// SetGlobalLogger sets the global logger instance directly
func SetGlobalLogger(logger Logger) {
	globalMutex.Lock()
	defer globalMutex.Unlock()
	globalLogger = logger
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() Logger {
	globalMutex.RLock()
	defer globalMutex.RUnlock()
	return globalLogger
}

// Global logger convenience functions

// Debug logs a debug message using the global logger
func Debug(ctx context.Context, msg string, fields ...Field) {
	if globalLogger != nil {
		globalLogger.Debug(ctx, msg, fields...)
	}
}

// Info logs an info message using the global logger
func Info(ctx context.Context, msg string, fields ...Field) {
	if globalLogger != nil {
		globalLogger.Info(ctx, msg, fields...)
	}
}

// Warn logs a warning message using the global logger
func Warn(ctx context.Context, msg string, fields ...Field) {
	if globalLogger != nil {
		globalLogger.Warn(ctx, msg, fields...)
	}
}

// Error logs an error message using the global logger
func Error(ctx context.Context, msg string, fields ...Field) {
	if globalLogger != nil {
		globalLogger.Error(ctx, msg, fields...)
	}
}

// Fatal logs a fatal message and exits using the global logger
func Fatal(ctx context.Context, msg string, fields ...Field) {
	if globalLogger != nil {
		globalLogger.Fatal(ctx, msg, fields...)
	}
}

// Convenience functions without context

// Debugf logs a debug message without context
func Debugf(msg string, fields ...Field) {
	if globalLogger != nil {
		globalLogger.Debugf(msg, fields...)
	}
}

// Infof logs an info message without context
func Infof(msg string, fields ...Field) {
	if globalLogger != nil {
		globalLogger.Infof(msg, fields...)
	}
}

// Warnf logs a warning message without context
func Warnf(msg string, fields ...Field) {
	if globalLogger != nil {
		globalLogger.Warnf(msg, fields...)
	}
}

// Errorf logs an error message without context
func Errorf(msg string, fields ...Field) {
	if globalLogger != nil {
		globalLogger.Errorf(msg, fields...)
	}
}

// WithFields returns a logger with additional fields
func WithFields(fields map[string]interface{}) Logger {
	if globalLogger != nil {
		return globalLogger.WithFields(fields)
	}
	return &NoOpLogger{}
}

// WithField returns a logger with an additional field
func WithField(key string, value interface{}) Logger {
	if globalLogger != nil {
		return globalLogger.WithField(key, value)
	}
	return &NoOpLogger{}
}

// WithError returns a logger with an error field
func WithError(err error) Logger {
	if globalLogger != nil {
		return globalLogger.WithError(err)
	}
	return &NoOpLogger{}
}

// Sync flushes any buffered log entries
func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}

// Enabled checks if the given level is enabled
func Enabled(level Level) bool {
	if globalLogger != nil {
		return globalLogger.Enabled(level)
	}
	return false
}

// NoOpLogger implements Logger interface with no-op operations
type NoOpLogger struct{}

func (l *NoOpLogger) Debug(ctx context.Context, msg string, fields ...Field) {}
func (l *NoOpLogger) Info(ctx context.Context, msg string, fields ...Field)  {}
func (l *NoOpLogger) Warn(ctx context.Context, msg string, fields ...Field)  {}
func (l *NoOpLogger) Error(ctx context.Context, msg string, fields ...Field) {}
func (l *NoOpLogger) DPanic(ctx context.Context, msg string, fields ...Field) {}
func (l *NoOpLogger) Panic(ctx context.Context, msg string, fields ...Field)  {}
func (l *NoOpLogger) Fatal(ctx context.Context, msg string, fields ...Field)  {}
func (l *NoOpLogger) Debugf(msg string, fields ...Field)                      {}
func (l *NoOpLogger) Infof(msg string, fields ...Field)                       {}
func (l *NoOpLogger) Warnf(msg string, fields ...Field)                       {}
func (l *NoOpLogger) Errorf(msg string, fields ...Field)                      {}
func (l *NoOpLogger) Enabled(level Level) bool                                { return false }
func (l *NoOpLogger) WithFields(fields map[string]interface{}) Logger         { return l }
func (l *NoOpLogger) WithField(key string, value interface{}) Logger          { return l }
func (l *NoOpLogger) WithError(err error) Logger                              { return l }
func (l *NoOpLogger) Sync() error                                             { return nil }
