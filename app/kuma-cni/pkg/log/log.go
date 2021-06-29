package log

import (
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

const logFilePath = "/var/log/kuma-cni.log"

var Log *logrus.Logger

func NewLogger() *logrus.Logger {
	if Log != nil {
		return Log
	}

	pathMap := lfshook.PathMap{
		logrus.PanicLevel: logFilePath,
		logrus.FatalLevel: logFilePath,
		logrus.ErrorLevel: logFilePath,
		logrus.WarnLevel:  logFilePath,
		logrus.InfoLevel:  logFilePath,
		logrus.DebugLevel: logFilePath,
		logrus.TraceLevel: logFilePath,
	}

	Log = logrus.New()
	Log.Hooks.Add(lfshook.NewHook(
		pathMap,
		&logrus.JSONFormatter{},
	))
	return Log
}

func init() {
	Log = NewLogger()
}
