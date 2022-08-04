package accesslogs

import (
	"errors"
)

func streamer(address string) (ReaderCloser, error) {
	return nil, errors.New("unsupported on Windows")
}
