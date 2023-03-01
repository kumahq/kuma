package xds

import (
	"fmt"

	envoy_log "github.com/envoyproxy/go-control-plane/pkg/log"
	"github.com/go-logr/logr"
)

func NewLogger(log logr.Logger) envoy_log.Logger {
	return &logger{log: log}
}

type logger struct {
	log logr.Logger
}

func (l logger) Debugf(format string, args ...interface{}) {
	l.log.V(1).Info(fmt.Sprintf(format, args...))
}

func (l logger) Warnf(format string, args ...interface{}) {
	l.log.V(1).Info(fmt.Sprintf(format, args...))
}

func (l logger) Infof(format string, args ...interface{}) {
	l.log.V(1).Info(fmt.Sprintf(format, args...))
}

func (l logger) Errorf(format string, args ...interface{}) {
	l.log.Error(fmt.Errorf(format, args...), "")
}
