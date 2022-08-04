//go:build !windows

package accesslogs

import (
	"os"
	"syscall"

	"github.com/pkg/errors"
)

func streamer(address string) (ReaderCloser, error) {
	err := os.Remove(address)
	if err != nil && !os.IsNotExist(err) {
		return nil, errors.Wrapf(err, "error removing existing fifo %s", address)
	}
	err = syscall.Mkfifo(address, 0666)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating fifo %s", address)
	}
	fd, err := os.OpenFile(address, os.O_CREATE, os.ModeNamedPipe)
	if err != nil {
		return nil, errors.Wrapf(err, "error opening fifo %s", address)
	}

	return fd, nil
}
