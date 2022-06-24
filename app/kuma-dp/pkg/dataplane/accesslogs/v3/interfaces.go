package v3

import (
	"io"

	"github.com/go-logr/logr"
)

// logHandler represents a contract between a log stream receiver and a log handler.
type logHandler interface {
	Handle(msg []byte) error
	io.Closer
}

// logSender represents a contract between a log handler and a log sender.
type logSender interface {
	Connect() error
	Send(entry []byte) error
	io.Closer
}

// logHandlerFactoryFunc represents a factory of log handler implementations.
type logHandlerFactoryFunc = func(log logr.Logger, address string) (logHandler, error)
