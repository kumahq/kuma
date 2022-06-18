package v3

import (
	"fmt"
	"strings"

	envoy_accesslog "github.com/envoyproxy/go-control-plane/envoy/service/accesslog/v3"
	"github.com/go-logr/logr"

	accesslog "github.com/kumahq/kuma/pkg/envoy/accesslog/v3"
)

func defaultHandler(log logr.Logger, msg *envoy_accesslog.StreamAccessLogsMessage) (logHandler, error) {
	parts := strings.SplitN(msg.GetIdentifier().GetLogName(), ";", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("log name %q has invalid format: expected %d components separated by ';', got %d", msg.GetIdentifier().GetLogName(), 2, len(parts))
	}
	address, formatString := parts[0], parts[1]

	format, err := accesslog.ParseFormat(formatString)
	if err != nil {
		return nil, err
	}

	sender := defaultSender(log, address)

	if err := sender.Connect(); err != nil {
		return nil, err
	}

	return &handler{
		format: format,
		sender: sender,
	}, nil
}

func defaultSender(log logr.Logger, address string) logSender {
	return &sender{
		log:     log,
		address: address,
	}
}
