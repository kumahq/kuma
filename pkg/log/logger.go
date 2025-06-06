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

	kds_middleware "github.com/kumahq/kuma/pkg/kds/middleware"
	"github.com/kumahq/kuma/pkg/multitenant"
	logger_extensions "github.com/kumahq/kuma/pkg/plugins/extensions/logger"
)

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

func newZapLoggerTo(destWriter io.Writer, level LogLevel, opts ...zap.Option) *zap.Logger {
	var lvl zap.AtomicLevel
	switch level {
	case OffLevel:
		return zap.NewNop()
	case DebugLevel:
		// The value we pass here is the most verbose level that
		// will end up being emitted through the `V(level int)`
		// accessor. Passing -10 ensures that levels up to `V(10)`
		// will work, which seems like plenty.
		lvl = zap.NewAtomicLevelAt(-10)
		opts = append(opts, zap.AddStacktrace(zap.ErrorLevel))
	default:
		lvl = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	encCfg := zap.NewDevelopmentEncoderConfig()
	enc := zapcore.NewConsoleEncoder(encCfg)
	sink := zapcore.AddSync(destWriter)
	opts = append(opts, zap.AddCallerSkip(1), zap.ErrorOutput(sink))
	return zap.New(zapcore.NewCore(&kube_log_zap.KubeAwareEncoder{Encoder: enc, Verbose: level == DebugLevel}, sink, lvl)).
		WithOptions(opts...)
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
