package v3

import (
	"io"
)

// logSender represents a contract between a log handler and a log sender.
type logSender interface {
	Connect() error
	Send(entry []byte) error
	io.Closer
}
