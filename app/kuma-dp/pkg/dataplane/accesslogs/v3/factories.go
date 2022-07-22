package v3

import (
	"github.com/go-logr/logr"
)

func defaultSender(log logr.Logger, address string) logSender {
	return &sender{
		log:     log,
		address: address,
	}
}
