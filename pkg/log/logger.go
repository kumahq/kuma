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

	"github.com/kumahq/kuma/pkg/multitenant"
)

type LogLevel int
type LogFormat int

const (
	OffLevel LogLevel = iota
	InfoLevel
	DebugLevel
)

const ( 
	Json LogFormat = iota
	Logfmt
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

func ParseLogFormat(text string) (LogFormat, error) {
	switch text {
	case "json":
		return Json, nil
	case "logfmt":
		return Logfmt, nil
	default: 
		return Logfmt, errors.Errorf("unkown log format %q", text)
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

func NewLogger(level LogLevel, format LogFormat) logr.Logger {
	return NewLoggerTo(os.Stderr, level, format)
}

func NewLoggerWithRotation(level LogLevel, format LogFormat, outputPath string, maxSize int, maxBackups int, maxAge int) logr.Logger {
	return NewLoggerTo(&lumberjack.Logger{
		Filename:   outputPath,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
	}, level, format)
}

func NewLoggerTo(destWriter io.Writer, level LogLevel, format LogFormat) logr.Logger {
	return zapr.NewLogger(newZapLoggerTo(destWriter, level, format))
}

func newZapLoggerTo(destWriter io.Writer, level LogLevel,format LogFormat, opts ...zap.Option) *zap.Logger {
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
	
	var enc zapcore.Encoder
	encCfg := zap.NewDevelopmentEncoderConfig()

	switch format {
	case Json:
		enc = zapcore.NewJSONEncoder(encCfg)
	case Logfmt:
		enc = zapcore.NewConsoleEncoder(encCfg)
	default:
		enc = zapcore.NewConsoleEncoder(encCfg)
	}
	
	sink := zapcore.AddSync(destWriter)
	opts = append(opts, zap.AddCallerSkip(1), zap.ErrorOutput(sink))
	return zap.New(zapcore.NewCore(&kube_log_zap.KubeAwareEncoder{Encoder: enc, Verbose: level == DebugLevel}, sink, lvl)).
		WithOptions(opts...)
}

const tenantLoggerKey = "tenantID"

// AddFieldsFromCtx will check if provided context contain tracing span and
// if the span is currently recording. If so, it will add trace_id and span_id
// to logged values. It will also add the tenant id to the logged values.
func AddFieldsFromCtx(logger logr.Logger, ctx context.Context) logr.Logger {
	if span := trace.SpanFromContext(ctx); span.IsRecording() {
		logger = logger.WithValues(
			"trace_id", span.SpanContext().TraceID(),
			"span_id", span.SpanContext().SpanID(),
		)
	}

	if tenantId, ok := multitenant.TenantFromCtx(ctx); ok {
		logger = logger.WithValues(tenantLoggerKey, tenantId)
	}

	return logger
}
