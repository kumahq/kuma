package accesslogs

import (
	"io"

	"github.com/go-logr/logr"

	envoy_accesslog "github.com/envoyproxy/go-control-plane/envoy/service/accesslog/v2"
)

// logHandler represents a contract between a log stream receiver and a log handler.
type logHandler interface {
	Handle(msg *envoy_accesslog.StreamAccessLogsMessage) error
	io.Closer
}

// logSender represents a contract between a log handler and a log sender.
type logSender interface {
	Connect() error
	Send(entry string) error
	io.Closer
}

// logHandlerFactoryFunc represents a factory of log handler implementations.
type logHandlerFactoryFunc = func(log logr.Logger, msg *envoy_accesslog.StreamAccessLogsMessage) (logHandler, error)
