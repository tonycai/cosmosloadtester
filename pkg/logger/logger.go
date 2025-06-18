package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/orijtech/cosmosloadtester/pkg/errors"
)

// LogLevel represents the logging level
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
	PanicLevel LogLevel = "panic"
)

// LogFormat represents the log output format
type LogFormat string

const (
	TextFormat LogFormat = "text"
	JSONFormat LogFormat = "json"
)

// Logger interface defines the logging contract
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	DebugWithFields(fields Fields, args ...interface{})
	
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	InfoWithFields(fields Fields, args ...interface{})
	
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	WarnWithFields(fields Fields, args ...interface{})
	
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	ErrorWithFields(fields Fields, args ...interface{})
	
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	FatalWithFields(fields Fields, args ...interface{})
	
	WithFields(fields Fields) Logger
	WithContext(ctx context.Context) Logger
	WithError(err error) Logger
	WithComponent(component string) Logger
}

// Fields represents structured log fields
type Fields map[string]interface{}

// Config holds logger configuration
type Config struct {
	Level      LogLevel  `json:"level" yaml:"level"`
	Format     LogFormat `json:"format" yaml:"format"`
	Output     string    `json:"output" yaml:"output"` // "stdout", "stderr", or file path
	MaxSize    int       `json:"max_size" yaml:"max_size"` // Max file size in MB
	MaxBackups int       `json:"max_backups" yaml:"max_backups"`
	MaxAge     int       `json:"max_age" yaml:"max_age"` // Max age in days
	Compress   bool      `json:"compress" yaml:"compress"`
	AddSource  bool      `json:"add_source" yaml:"add_source"`
}

// DefaultConfig returns a default logger configuration
func DefaultConfig() *Config {
	return &Config{
		Level:      InfoLevel,
		Format:     TextFormat,
		Output:     "stdout",
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
		AddSource:  false,
	}
}

// LoadTestLogger implements the Logger interface using logrus
type LoadTestLogger struct {
	logger    *logrus.Logger
	entry     *logrus.Entry
	config    *Config
	mu        sync.RWMutex
	component string
}

// NewLogger creates a new logger instance
func NewLogger(config *Config) (*LoadTestLogger, error) {
	if config == nil {
		config = DefaultConfig()
	}

	logger := logrus.New()
	
	// Set log level
	level, err := logrus.ParseLevel(string(config.Level))
	if err != nil {
		return nil, errors.NewConfigError(errors.ErrCodeInvalidConfig, 
			fmt.Sprintf("invalid log level: %s", config.Level))
	}
	logger.SetLevel(level)

	// Set formatter
	switch config.Format {
	case JSONFormat:
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
				logrus.FieldKeyFunc:  "caller",
			},
		})
	case TextFormat:
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
			DisableColors:   false,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "time",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "msg",
			},
		})
	default:
		return nil, errors.NewConfigError(errors.ErrCodeInvalidConfig,
			fmt.Sprintf("invalid log format: %s", config.Format))
	}

	// Set output
	output, err := getLogOutput(config.Output)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeConfig, 
			errors.ErrCodeInvalidConfig, "failed to set log output")
	}
	logger.SetOutput(output)

	// Add caller information if requested
	if config.AddSource {
		logger.SetReportCaller(true)
	}

	return &LoadTestLogger{
		logger: logger,
		entry:  logrus.NewEntry(logger),
		config: config,
	}, nil
}

// NewLoggerWithDefaults creates a logger with default configuration
func NewLoggerWithDefaults() *LoadTestLogger {
	logger, err := NewLogger(DefaultConfig())
	if err != nil {
		// Fallback to basic logrus logger
		fallback := logrus.New()
		return &LoadTestLogger{
			logger: fallback,
			entry:  logrus.NewEntry(fallback),
			config: DefaultConfig(),
		}
	}
	return logger
}

// Debug logs a debug message
func (l *LoadTestLogger) Debug(args ...interface{}) {
	l.entry.Debug(args...)
}

// Debugf logs a formatted debug message
func (l *LoadTestLogger) Debugf(format string, args ...interface{}) {
	l.entry.Debugf(format, args...)
}

// DebugWithFields logs a debug message with structured fields
func (l *LoadTestLogger) DebugWithFields(fields Fields, args ...interface{}) {
	l.entry.WithFields(logrus.Fields(fields)).Debug(args...)
}

// Info logs an info message
func (l *LoadTestLogger) Info(args ...interface{}) {
	l.entry.Info(args...)
}

// Infof logs a formatted info message
func (l *LoadTestLogger) Infof(format string, args ...interface{}) {
	l.entry.Infof(format, args...)
}

// InfoWithFields logs an info message with structured fields
func (l *LoadTestLogger) InfoWithFields(fields Fields, args ...interface{}) {
	l.entry.WithFields(logrus.Fields(fields)).Info(args...)
}

// Warn logs a warning message
func (l *LoadTestLogger) Warn(args ...interface{}) {
	l.entry.Warn(args...)
}

// Warnf logs a formatted warning message
func (l *LoadTestLogger) Warnf(format string, args ...interface{}) {
	l.entry.Warnf(format, args...)
}

// WarnWithFields logs a warning message with structured fields
func (l *LoadTestLogger) WarnWithFields(fields Fields, args ...interface{}) {
	l.entry.WithFields(logrus.Fields(fields)).Warn(args...)
}

// Error logs an error message
func (l *LoadTestLogger) Error(args ...interface{}) {
	l.entry.Error(args...)
}

// Errorf logs a formatted error message
func (l *LoadTestLogger) Errorf(format string, args ...interface{}) {
	l.entry.Errorf(format, args...)
}

// ErrorWithFields logs an error message with structured fields
func (l *LoadTestLogger) ErrorWithFields(fields Fields, args ...interface{}) {
	l.entry.WithFields(logrus.Fields(fields)).Error(args...)
}

// Fatal logs a fatal message and exits
func (l *LoadTestLogger) Fatal(args ...interface{}) {
	l.entry.Fatal(args...)
}

// Fatalf logs a formatted fatal message and exits
func (l *LoadTestLogger) Fatalf(format string, args ...interface{}) {
	l.entry.Fatalf(format, args...)
}

// FatalWithFields logs a fatal message with structured fields and exits
func (l *LoadTestLogger) FatalWithFields(fields Fields, args ...interface{}) {
	l.entry.WithFields(logrus.Fields(fields)).Fatal(args...)
}

// WithFields creates a new logger entry with the given fields
func (l *LoadTestLogger) WithFields(fields Fields) Logger {
	return &LoadTestLogger{
		logger:    l.logger,
		entry:     l.entry.WithFields(logrus.Fields(fields)),
		config:    l.config,
		component: l.component,
	}
}

// WithContext creates a new logger entry with context information
func (l *LoadTestLogger) WithContext(ctx context.Context) Logger {
	fields := extractContextFields(ctx)
	return l.WithFields(fields)
}

// WithError creates a new logger entry with error information
func (l *LoadTestLogger) WithError(err error) Logger {
	fields := Fields{"error": err.Error()}
	
	// Add structured error information if it's a LoadTestError
	if loadTestErr, ok := err.(*errors.LoadTestError); ok {
		fields["error_type"] = loadTestErr.Type
		fields["error_code"] = loadTestErr.Code
		fields["component"] = loadTestErr.Component
		if loadTestErr.Context != nil {
			for k, v := range loadTestErr.Context {
				fields[fmt.Sprintf("error_context_%s", k)] = v
			}
		}
	}
	
	return l.WithFields(fields)
}

// WithComponent creates a new logger entry with component information
func (l *LoadTestLogger) WithComponent(component string) Logger {
	return &LoadTestLogger{
		logger:    l.logger,
		entry:     l.entry.WithField("component", component),
		config:    l.config,
		component: component,
	}
}

// LogError logs an error with appropriate level and context
func (l *LoadTestLogger) LogError(err error) {
	if err == nil {
		return
	}

	logger := l.WithError(err)
	
	// Use different log levels based on error type
	if loadTestErr, ok := err.(*errors.LoadTestError); ok {
		switch loadTestErr.Type {
		case errors.ErrorTypeValidation, errors.ErrorTypeConfig:
			logger.Warn("Configuration or validation error occurred")
		case errors.ErrorTypeNetwork, errors.ErrorTypeConnection, errors.ErrorTypeTimeout:
			logger.Error("Network or connectivity error occurred")
		case errors.ErrorTypeLoadTest, errors.ErrorTypeTransaction, errors.ErrorTypeBroadcast:
			logger.Error("Load test execution error occurred")
		case errors.ErrorTypeInternal:
			logger.Error("Internal error occurred")
		default:
			logger.Error("Unknown error occurred")
		}
	} else {
		logger.Error("Error occurred")
	}
}

// Utility functions

func getLogOutput(output string) (io.Writer, error) {
	switch output {
	case "stdout":
		return os.Stdout, nil
	case "stderr":
		return os.Stderr, nil
	default:
		// Assume it's a file path
		dir := filepath.Dir(output)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
		
		file, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		return file, nil
	}
}

func extractContextFields(ctx context.Context) Fields {
	fields := Fields{}
	
	// Extract common context values
	if requestID := ctx.Value("request_id"); requestID != nil {
		fields["request_id"] = requestID
	}
	
	if userID := ctx.Value("user_id"); userID != nil {
		fields["user_id"] = userID
	}
	
	if traceID := ctx.Value("trace_id"); traceID != nil {
		fields["trace_id"] = traceID
	}
	
	return fields
}

// Global logger instance
var (
	globalLogger Logger
	loggerMu     sync.RWMutex
)

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(logger Logger) {
	loggerMu.Lock()
	defer loggerMu.Unlock()
	globalLogger = logger
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() Logger {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	
	if globalLogger == nil {
		globalLogger = NewLoggerWithDefaults()
	}
	
	return globalLogger
}

// Convenience functions for global logger

// Debug logs a debug message using the global logger
func Debug(args ...interface{}) {
	GetGlobalLogger().Debug(args...)
}

// Debugf logs a formatted debug message using the global logger
func Debugf(format string, args ...interface{}) {
	GetGlobalLogger().Debugf(format, args...)
}

// Info logs an info message using the global logger
func Info(args ...interface{}) {
	GetGlobalLogger().Info(args...)
}

// Infof logs a formatted info message using the global logger
func Infof(format string, args ...interface{}) {
	GetGlobalLogger().Infof(format, args...)
}

// Warn logs a warning message using the global logger
func Warn(args ...interface{}) {
	GetGlobalLogger().Warn(args...)
}

// Warnf logs a formatted warning message using the global logger
func Warnf(format string, args ...interface{}) {
	GetGlobalLogger().Warnf(format, args...)
}

// Error logs an error message using the global logger
func Error(args ...interface{}) {
	GetGlobalLogger().Error(args...)
}

// Errorf logs a formatted error message using the global logger
func Errorf(format string, args ...interface{}) {
	GetGlobalLogger().Errorf(format, args...)
}

// Fatal logs a fatal message using the global logger and exits
func Fatal(args ...interface{}) {
	GetGlobalLogger().Fatal(args...)
}

// Fatalf logs a formatted fatal message using the global logger and exits
func Fatalf(format string, args ...interface{}) {
	GetGlobalLogger().Fatalf(format, args...)
}

// WithFields creates a logger with fields using the global logger
func WithFields(fields Fields) Logger {
	return GetGlobalLogger().WithFields(fields)
}

// WithError creates a logger with error using the global logger
func WithError(err error) Logger {
	return GetGlobalLogger().WithError(err)
}

// WithComponent creates a logger with component using the global logger
func WithComponent(component string) Logger {
	return GetGlobalLogger().WithComponent(component)
}

// LogError logs an error using the global logger
func LogError(err error) {
	if logger, ok := GetGlobalLogger().(*LoadTestLogger); ok {
		logger.LogError(err)
	} else {
		GetGlobalLogger().WithError(err).Error("Error occurred")
	}
}

// Helper function to get caller information
func getCaller() string {
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		return "unknown"
	}
	
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return fmt.Sprintf("%s:%d", filepath.Base(file), line)
	}
	
	fnName := fn.Name()
	parts := strings.Split(fnName, "/")
	if len(parts) > 0 {
		fnName = parts[len(parts)-1]
	}
	
	return fmt.Sprintf("%s:%d:%s", filepath.Base(file), line, fnName)
} 