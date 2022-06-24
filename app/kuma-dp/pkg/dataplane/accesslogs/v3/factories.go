package v3

import (
	"github.com/go-logr/logr"
)

func defaultHandler(log logr.Logger, address string) (logHandler, error) {
	sender := defaultSender(log, address)

	if err := sender.Connect(); err != nil {
		return nil, err
	}

	return &handler{
		sender: sender,
	}, nil
}

func defaultSender(log logr.Logger, address string) logSender {
	return &sender{
		log:     log,
		address: address,
	}
}
