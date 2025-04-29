// Package logger provides a centralized logging system using Zap
package logger

import (
	"context"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Log is the global logger
	Log *zap.Logger

	// SugaredLog is the global sugared logger (with formatting capabilities)
	SugaredLog *zap.SugaredLogger

	// once ensures the logger is initialized only once
	once sync.Once
)

// Config contains logger configuration
type Config struct {
	// Development puts the logger in development mode, which changes the
	// behavior of DPanicLevel and some other minor details
	Development bool

	// Level is the minimum enabled logging level
	Level string

	// OutputPaths is a list of URLs or file paths to write logging output to
	OutputPaths []string
}

// Setup initializes the logger with the given configuration
func Setup(cfg Config) {
	once.Do(func() {
		// Parse log level
		logLevel := zap.InfoLevel
		if err := logLevel.UnmarshalText([]byte(cfg.Level)); err != nil {
			// Default to info level on parsing error
			logLevel = zap.InfoLevel
		}

		// Set up encoder config
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.TimeKey = "timestamp"
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		// Set reasonable defaults for output paths
		outputPaths := cfg.OutputPaths
		if len(outputPaths) == 0 {
			outputPaths = []string{"stdout"}
		}

		// Build logger config
		config := zap.Config{
			Level:             zap.NewAtomicLevelAt(logLevel),
			Development:       cfg.Development,
			DisableCaller:     false,
			DisableStacktrace: false,
			Sampling:          nil, // Disable sampling
			Encoding:          "json",
			EncoderConfig:     encoderConfig,
			OutputPaths:       outputPaths,
			ErrorOutputPaths:  []string{"stderr"},
		}

		var err error
		Log, err = config.Build()
		if err != nil {
			// If we can't build the logger, fall back to a basic logger
			Log = zap.New(zapcore.NewCore(
				zapcore.NewJSONEncoder(encoderConfig),
				zapcore.AddSync(os.Stdout),
				zap.InfoLevel,
			))
		}

		// Create sugared logger
		SugaredLog = Log.Sugar()

		// Make sure we clean up
		zap.ReplaceGlobals(Log)

		Log.Info("Logger initialized",
			zap.String("level", logLevel.String()),
			zap.Bool("development", cfg.Development))
	})
}

// With returns a logger with additional fields added to the logging context
func With(fields ...zapcore.Field) *zap.Logger {
	return Log.With(fields...)
}

// WithContext returns a logger with fields from the given context
func WithContext(ctx context.Context) *zap.Logger {
	// Example of extracting request ID from context (could be customized)
	// if requestID, ok := ctx.Value("request_id").(string); ok {
	//     return Log.With(zap.String("request_id", requestID))
	// }
	return Log
}

// Sync flushes any buffered log entries
func Sync() error {
	return Log.Sync()
}

// Fields is an alias for zapcore.Field
type Field = zapcore.Field

// Common field constructors
var (
	Any        = zap.Any
	Bool       = zap.Bool
	Duration   = zap.Duration
	Float64    = zap.Float64
	Int        = zap.Int
	Int64      = zap.Int64
	String     = zap.String
	Stringer   = zap.Stringer
	Time       = zap.Time
	Uint       = zap.Uint
	Uint64     = zap.Uint64
	Error      = zap.Error
	NamedError = zap.NamedError
	Stack      = zap.Stack
	StackSkip  = zap.StackSkip
	Binary     = zap.Binary
	ByteString = zap.ByteString
	Namespace  = zap.Namespace
	Reflect    = zap.Reflect
)
