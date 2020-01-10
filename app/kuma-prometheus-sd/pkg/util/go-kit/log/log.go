package log

import (
	"github.com/go-logr/logr"
)

func NewLogger(log logr.Logger, component string) *logger {
	return &logger{
		log:       log,
		component: component,
	}
}

type logger struct {
	log       logr.Logger
	component string
}

func (l *logger) Log(keyvals ...interface{}) error {
	l.log.Info(l.component, keyvals...)
	return nil
}
