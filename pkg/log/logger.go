package log

import (
	"context"
	"io"
	"os"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	kube_log_zap "sigs.k8s.io/controller-runtime/pkg/log/zap"
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

func NewContext(ctx context.Context, logger logr.Logger) context.Context {
	return logr.NewContext(ctx, logger)
}

func FromContext(ctx context.Context) (logr.Logger, error) {
	logger, err := logr.FromContext(ctx)
	if err != nil {
		return logr.Logger{}, errors.Wrap(err, "could not extract logger from the context")
	}

	return logger, nil
}

func FromContextOrDefault(
	ctx context.Context,
	defaultLogger logr.Logger,
) logr.Logger {
	logger, err := logr.FromContext(ctx)
	if err != nil {
		return defaultLogger
	}

	return logger
}

func FromContextWithNameAndOptionalValues(
	ctx context.Context,
	name string,
	values ...interface{},
) (logr.Logger, error) {
	logger, err := FromContext(ctx)
	if err != nil {
		return logr.Logger{}, err
	}

	return logger.WithName(name).WithValues(values...), nil
}

func FromContextWithNameAndOptionalValuesOrDefault(
	ctx context.Context,
	defaultLogger logr.Logger,
	name string,
	values ...interface{},
) logr.Logger {
	logger, err := logr.FromContext(ctx)
	if err != nil {
		logger = defaultLogger
	}

	return logger.WithName(name).WithValues(values...)
}
