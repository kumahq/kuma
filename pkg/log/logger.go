package log

import (
	"context"
	"io"
	"os"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	kube_log_zap "sigs.k8s.io/controller-runtime/pkg/log/zap"

	kds_middleware "github.com/kumahq/kuma/v2/pkg/kds/middleware"
	"github.com/kumahq/kuma/v2/pkg/multitenant"
	logger_extensions "github.com/kumahq/kuma/v2/pkg/plugins/extensions/logger"
)

// defaultAtomicLevel is the global atomic level used by NewLoggerWithGlobalLevel.
// This allows dynamic log level changes without replacing the logger instance.
var defaultAtomicLevel = zap.NewAtomicLevel()

type LogLevel int

const (
	OffLevel LogLevel = iota
	InfoLevel
	DebugLevel
)

func (l LogLevel) String() string {
	switch l {
	case OffLevel:
		return "off"
	case InfoLevel:
		return "info"
	case DebugLevel:
		return "debug"
	default:
		return "unknown"
	}
}

func ParseLogLevel(text string) (LogLevel, error) {
	switch text {
	case "off":
		return OffLevel, nil
	case "info":
		return InfoLevel, nil
	case "debug":
		return DebugLevel, nil
	default:
		return OffLevel, errors.Errorf("unknown log level %q", text)
	}
}

func NewLogger(level LogLevel) logr.Logger {
	return NewLoggerTo(os.Stderr, level)
}

func NewLoggerWithRotation(level LogLevel, outputPath string, maxSize int, maxBackups int, maxAge int) logr.Logger {
	return NewLoggerTo(&lumberjack.Logger{
		Filename:   outputPath,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
	}, level)
}

func NewLoggerTo(destWriter io.Writer, level LogLevel) logr.Logger {
	return zapr.NewLogger(newZapLoggerTo(destWriter, level))
}

// NewLoggerWithGlobalLevel creates a logger that uses the global atomic level.
// The log level can be changed dynamically using SetGlobalLogLevel without
// replacing the logger instance. This is useful for early initialization
// (e.g., in init() functions) where the final log level is not yet known.
func NewLoggerWithGlobalLevel() logr.Logger {
	return zapr.NewLogger(newZapLoggerWithGlobalLevel(os.Stderr))
}

// SetGlobalLogLevel updates the global atomic log level used by loggers
// created with NewLoggerWithGlobalLevel. This allows changing the log level
// of all such loggers without replacing the logger instances.
func SetGlobalLogLevel(level LogLevel) {
	switch level {
	case OffLevel:
		// Set to a very high level to effectively disable logging
		defaultAtomicLevel.SetLevel(zapcore.Level(100))
	case DebugLevel:
		// The value we pass here is the most verbose level that
		// will end up being emitted through the `V(level int)`
		// accessor. Passing -10 ensures that levels up to `V(10)`
		// will work, which seems like plenty.
		defaultAtomicLevel.SetLevel(zapcore.Level(-10))
	case InfoLevel:
		defaultAtomicLevel.SetLevel(zapcore.InfoLevel)
	}
}

func newZapLoggerTo(destWriter io.Writer, level LogLevel, opts ...zap.Option) *zap.Logger {
	if level == OffLevel {
		return zap.NewNop()
	}

	// Create a new AtomicLevel for this logger
	var lvl zap.AtomicLevel
	var verbose bool

	switch level {
	case DebugLevel:
		// The value we pass here is the most verbose level that
		// will end up being emitted through the `V(level int)`
		// accessor. Passing -10 ensures that levels up to `V(10)`
		// will work, which seems like plenty.
		lvl = zap.NewAtomicLevelAt(-10)
		verbose = true
	default:
		lvl = zap.NewAtomicLevelAt(zap.InfoLevel)
		verbose = false
	}

	return buildZapLogger(destWriter, lvl, verbose, opts...)
}

func newZapLoggerWithGlobalLevel(destWriter io.Writer, opts ...zap.Option) *zap.Logger {
	// defaultAtomicLevel starts at InfoLevel (zap.NewAtomicLevel() default).
	// The level can be changed via SetGlobalLogLevel() before or after this call.

	// Note: verbose is determined at creation time based on the current level.
	// If the level changes later via SetGlobalLogLevel(), the verbose flag
	// (affecting KubeAwareEncoder.Verbose and stacktrace behavior) will NOT
	// update. For kumactl's use case, this is acceptable since we only call
	// SetGlobalLogLevel() once during flag parsing, before any logging occurs.
	verbose := defaultAtomicLevel.Level() <= zapcore.Level(-10)

	return buildZapLogger(destWriter, defaultAtomicLevel, verbose, opts...)
}

// buildZapLogger is the common implementation for creating zap loggers.
// It takes an AtomicLevel (either per-logger or global) and constructs
// the logger with the standard Kuma configuration.
//
// Parameters:
//   - destWriter: Where log output will be written.
//   - lvl: AtomicLevel that controls the log level (can be shared across loggers).
//   - verbose: If true, enables verbose mode for KubeAwareEncoder and adds
//     stacktraces at ErrorLevel. This is fixed at creation time.
//   - opts: Additional zap options to apply to the logger.
func buildZapLogger(
	destWriter io.Writer,
	lvl zap.AtomicLevel,
	verbose bool,
	opts ...zap.Option,
) *zap.Logger {
	encCfg := zap.NewDevelopmentEncoderConfig()
	enc := zapcore.NewConsoleEncoder(encCfg)
	sink := zapcore.AddSync(destWriter)

	// Add standard options
	opts = append(opts, zap.AddCallerSkip(1), zap.ErrorOutput(sink))
	if verbose {
		opts = append(opts, zap.AddStacktrace(zap.ErrorLevel))
	}

	encoder := &kube_log_zap.KubeAwareEncoder{
		Encoder: enc,
		Verbose: verbose,
	}

	return zap.New(zapcore.NewCore(encoder, sink, lvl)).WithOptions(opts...)
}

const (
	TenantLoggerKey   = "tenantID"
	StreamIDLoggerKey = "streamID"
)

// AddFieldsFromCtx will check if provided context contain tracing span and
// if the span is currently recording. If so, it will call spanLogValuesProcessor
// function if it's also present in the context. If not it will add trace_id
// and span_id to logged values. It will also add the tenant id to the logged
// values.
func AddFieldsFromCtx(
	logger logr.Logger,
	ctx context.Context,
	extensions context.Context,
) logr.Logger {
	if tenantId, ok := multitenant.TenantFromCtx(ctx); ok {
		logger = logger.WithValues(TenantLoggerKey, tenantId)
	}

	if streamID, ok := kds_middleware.StreamIDFromCtx(ctx); ok {
		logger = logger.WithValues(StreamIDLoggerKey, streamID)
	}

	return addSpanValuesToLogger(logger, ctx, extensions)
}

func addSpanValuesToLogger(
	logger logr.Logger,
	ctx context.Context,
	extensions context.Context,
) logr.Logger {
	if span := trace.SpanFromContext(ctx); span.IsRecording() {
		if fn, ok := logger_extensions.FromSpanLogValuesProcessorContext(extensions); ok {
			return logger.WithValues(fn(span)...)
		}

		return logger.WithValues(
			"trace_id", span.SpanContext().TraceID(),
			"span_id", span.SpanContext().SpanID(),
		)
	}

	return logger
}
