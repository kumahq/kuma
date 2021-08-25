package log

import (
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

const logFilePath = "/var/log/kuma-cni.log"

var Log *logrus.Logger

type logger struct {
	Log     *logrus.Logger
	logFile *os.File
}

func NewLogger() *logger {
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		panic(fmt.Errorf("unable to open %s: %v", logFilePath, err))
	}

	result := &logger{
		Log:     logrus.New(),
		logFile: logFile,
	}

	result.Log.SetOutput(io.MultiWriter(os.Stderr, logFile))
	result.Log.SetFormatter(&logrus.JSONFormatter{})

	return result
}

func (l *logger) Exit(code int) {
	if l.logFile != nil {
		l.logFile.Close()
	}
	l.Log.Exit(code)
}
