package clog

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Level represents the logging level
type Level int8

const (
	DebugLevel Level = iota - 1
	InfoLevel
	WarnLevel
	ErrorLevel
	DPanicLevel
	PanicLevel
	FatalLevel
)

// ParseLevel parses a level string
func ParseLevel(s string) Level {
	switch strings.ToLower(s) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "dpanic":
		return DPanicLevel
	case "panic":
		return PanicLevel
	case "fatal":
		return FatalLevel
	default:
		return InfoLevel
	}
}

func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case DPanicLevel:
		return "dpanic"
	case PanicLevel:
		return "panic"
	case FatalLevel:
		return "fatal"
	default:
		return "info"
	}
}

// Field represents a key-value pair for structured logging
type Field struct {
	Key   string
	Value interface{}
}

// Logger defines the core logging interface
type Logger interface {
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, fields ...Field)
	DPanic(ctx context.Context, msg string, fields ...Field)
	Panic(ctx context.Context, msg string, fields ...Field)
	Fatal(ctx context.Context, msg string, fields ...Field)

	// Convenience methods without context
	Debugf(msg string, fields ...Field)
	Infof(msg string, fields ...Field)
	Warnf(msg string, fields ...Field)
	Errorf(msg string, fields ...Field)

	Enabled(level Level) bool

	WithFields(fields map[string]interface{}) Logger
	WithField(key string, value interface{}) Logger
	WithError(err error) Logger

	Sync() error
}

// Config holds the configuration for logger
type Config struct {
	Level             string          `yaml:"level" json:"level"`
	Format            string          `yaml:"format" json:"format"` // json, console
	DisableCaller     bool            `yaml:"disable_caller" json:"disable_caller"`
	OutputPaths       []string        `yaml:"output_paths" json:"output_paths"`
	Rotation          *RotationConfig `yaml:"rotation" json:"rotation"`
}

// RotationConfig holds log rotation settings
type RotationConfig struct {
	MaxSize    int  `yaml:"max_size" json:"max_size"`       // MB
	MaxBackups int  `yaml:"max_backups" json:"max_backups"` // number of backups
	MaxAge     int  `yaml:"max_age" json:"max_age"`         // days
	Compress   bool `yaml:"compress" json:"compress"`
}

// DefaultConfig returns the default logger configuration
func DefaultConfig() *Config {
	return &Config{
		Level:         "info",
		Format:        "console",
		DisableCaller: false,
		OutputPaths:   []string{"stdout"},
		Rotation: &RotationConfig{
			MaxSize:    10,
			MaxBackups: 3,
			MaxAge:     7,
			Compress:   true,
		},
	}
}

// ConfigFromSettings creates config from AIM settings
func ConfigFromSettings(logLevel string) *Config {
	config := DefaultConfig()
	if logLevel != "" {
		config.Level = logLevel
	}
	return config
}

// ZapLoggerImpl implements Logger interface using zap
type ZapLoggerImpl struct {
	logger     *zap.Logger
	config     *Config
	callerSkip int
}

// NewLogger creates a new zap-based logger
func NewLogger(config *Config) (Logger, error) {
	return NewLoggerWithCallerSkip(config, 0)
}

// NewLoggerWithCallerSkip creates a new zap-based logger with caller skip
func NewLoggerWithCallerSkip(config *Config, callerSkip int) (Logger, error) {
	if config == nil {
		config = DefaultConfig()
	}

	level, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		return nil, err
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   customCallerEncoder,
	}

	var encoder zapcore.Encoder
	switch config.Format {
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	case "console":
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	writeSyncers := getWriteSyncers(config.OutputPaths, config.Rotation)
	multiWriteSyncer := zapcore.NewMultiWriteSyncer(writeSyncers...)
	core := zapcore.NewCore(encoder, multiWriteSyncer, level)

	options := []zap.Option{}
	if !config.DisableCaller {
		options = append(options, zap.AddCaller())
		if callerSkip > 0 {
			options = append(options, zap.AddCallerSkip(callerSkip))
		}
	}

	logger := zap.New(core, options...)

	return &ZapLoggerImpl{
		logger:     logger,
		config:     config,
		callerSkip: callerSkip,
	}, nil
}

// customCallerEncoder encodes caller information with 2 path segments
func customCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	if !caller.Defined {
		enc.AppendString("undefined")
		return
	}

	parts := strings.Split(caller.File, "/")
	var pathSegments []string

	if len(parts) >= 2 {
		pathSegments = parts[len(parts)-2:]
	} else {
		pathSegments = parts
	}

	trimmed := strings.Join(pathSegments, "/")
	enc.AppendString(fmt.Sprintf("%s:%d", trimmed, caller.Line))
}

// getWriteSyncers converts output paths to WriteSyncers
func getWriteSyncers(paths []string, rotationConfig *RotationConfig) []zapcore.WriteSyncer {
	var syncers []zapcore.WriteSyncer
	for _, path := range paths {
		switch path {
		case "stdout":
			syncers = append(syncers, zapcore.AddSync(os.Stdout))
		case "stderr":
			syncers = append(syncers, zapcore.AddSync(os.Stderr))
		default:
			if rotationConfig != nil && isLogFile(path) {
				lumberjackLogger := &lumberjack.Logger{
					Filename:   path,
					MaxSize:    rotationConfig.MaxSize,
					MaxBackups: rotationConfig.MaxBackups,
					MaxAge:     rotationConfig.MaxAge,
					Compress:   rotationConfig.Compress,
				}
				syncers = append(syncers, zapcore.AddSync(lumberjackLogger))
			} else {
				if file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644); err == nil {
					syncers = append(syncers, zapcore.AddSync(file))
				}
			}
		}
	}
	return syncers
}

// isLogFile checks if the path is a file path
func isLogFile(path string) bool {
	return filepath.IsAbs(path) || (path != "stdout" && path != "stderr" && filepath.Ext(path) != "")
}

// Logger implementation methods
func (l *ZapLoggerImpl) Debug(ctx context.Context, msg string, fields ...Field) {
	l.log(zapcore.DebugLevel, msg, fields...)
}

func (l *ZapLoggerImpl) Info(ctx context.Context, msg string, fields ...Field) {
	l.log(zapcore.InfoLevel, msg, fields...)
}

func (l *ZapLoggerImpl) Warn(ctx context.Context, msg string, fields ...Field) {
	l.log(zapcore.WarnLevel, msg, fields...)
}

func (l *ZapLoggerImpl) Error(ctx context.Context, msg string, fields ...Field) {
	l.log(zapcore.ErrorLevel, msg, fields...)
}

func (l *ZapLoggerImpl) DPanic(ctx context.Context, msg string, fields ...Field) {
	l.log(zapcore.DPanicLevel, msg, fields...)
}

func (l *ZapLoggerImpl) Panic(ctx context.Context, msg string, fields ...Field) {
	l.log(zapcore.PanicLevel, msg, fields...)
}

func (l *ZapLoggerImpl) Fatal(ctx context.Context, msg string, fields ...Field) {
	l.log(zapcore.FatalLevel, msg, fields...)
}

// Convenience methods without context
func (l *ZapLoggerImpl) Debugf(msg string, fields ...Field) {
	l.log(zapcore.DebugLevel, msg, fields...)
}

func (l *ZapLoggerImpl) Infof(msg string, fields ...Field) {
	l.log(zapcore.InfoLevel, msg, fields...)
}

func (l *ZapLoggerImpl) Warnf(msg string, fields ...Field) {
	l.log(zapcore.WarnLevel, msg, fields...)
}

func (l *ZapLoggerImpl) Errorf(msg string, fields ...Field) {
	l.log(zapcore.ErrorLevel, msg, fields...)
}

// log is the internal logging method
func (l *ZapLoggerImpl) log(level zapcore.Level, msg string, fields ...Field) {
	zapFields := ToZapFields(fields)

	switch level {
	case zapcore.DebugLevel:
		l.logger.Debug(msg, zapFields...)
	case zapcore.InfoLevel:
		l.logger.Info(msg, zapFields...)
	case zapcore.WarnLevel:
		l.logger.Warn(msg, zapFields...)
	case zapcore.ErrorLevel:
		l.logger.Error(msg, zapFields...)
	case zapcore.DPanicLevel:
		l.logger.DPanic(msg, zapFields...)
	case zapcore.PanicLevel:
		l.logger.Panic(msg, zapFields...)
	case zapcore.FatalLevel:
		l.logger.Fatal(msg, zapFields...)
	}
}

func (l *ZapLoggerImpl) Enabled(level Level) bool {
	return l.logger.Core().Enabled(l.convertLevel(level))
}

func (l *ZapLoggerImpl) convertLevel(level Level) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case DPanicLevel:
		return zapcore.DPanicLevel
	case PanicLevel:
		return zapcore.PanicLevel
	case FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func (l *ZapLoggerImpl) WithFields(fields map[string]interface{}) Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return &ZapLoggerImpl{
		logger:     l.logger.With(zapFields...),
		config:     l.config,
		callerSkip: l.callerSkip,
	}
}

func (l *ZapLoggerImpl) WithField(key string, value interface{}) Logger {
	return &ZapLoggerImpl{
		logger:     l.logger.With(zap.Any(key, value)),
		config:     l.config,
		callerSkip: l.callerSkip,
	}
}

func (l *ZapLoggerImpl) WithError(err error) Logger {
	return l.WithField("error", err)
}

func (l *ZapLoggerImpl) Sync() error {
	return l.logger.Sync()
}

// Helper functions for creating fields
func String(key, val string) Field {
	return Field{Key: key, Value: val}
}

func Int(key string, val int) Field {
	return Field{Key: key, Value: val}
}

func Int64(key string, val int64) Field {
	return Field{Key: key, Value: val}
}

func Bool(key string, val bool) Field {
	return Field{Key: key, Value: val}
}

func Duration(key string, val time.Duration) Field {
	return Field{Key: key, Value: val}
}

func Time(key string, val time.Time) Field {
	return Field{Key: key, Value: val}
}

func Any(key string, val interface{}) Field {
	return Field{Key: key, Value: val}
}

func Err(err error) Field {
	return Field{Key: "error", Value: err}
}

// Convert our Field to zap.Field
func (f Field) ToZapField() zap.Field {
	return zap.Any(f.Key, f.Value)
}

// Convert multiple fields to zap fields
func ToZapFields(fields []Field) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = field.ToZapField()
	}
	return zapFields
}
