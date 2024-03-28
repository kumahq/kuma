package channels

import "github.com/labstack/gommon/log"

func NonBlockingErrorWrite(ch chan error, err error) {
	select {
	case ch <- err:
	default:
		log.Error(err, "Failed to write error to closed channel")
	}
}
