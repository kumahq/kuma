package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

const (
	logLocation = "/tmp/kuma-cni.log"
)

var (
	Default *zap.Logger
)

func init() {
	core := defaultConfig(zapcore.DebugLevel)
	Default = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

func InitLogger(logLevel string) {
	core := defaultConfig(mapLogLevel(logLevel))
	Default = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

func mapLogLevel(logLevel string) zapcore.Level {
	switch logLevel {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "error":
		return zapcore.ErrorLevel
	case "warn":
		return zapcore.WarnLevel
	default:
		return zapcore.InfoLevel
	}
}

func defaultConfig(logLevel zapcore.Level) zapcore.Core {
	config := zap.NewProductionEncoderConfig()
	fileEncoder := zapcore.NewJSONEncoder(config)
	logFile, _ := os.OpenFile(logLocation, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	stderrSyncer := zapcore.Lock(os.Stderr)
	writer := zapcore.AddSync(logFile)
	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, writer, logLevel),
		zapcore.NewCore(fileEncoder, stderrSyncer, logLevel),
	)
	return core
}
