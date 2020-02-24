package accesslogs

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/go-logr/logr"

	envoy_accesslog "github.com/envoyproxy/go-control-plane/envoy/service/accesslog/v2"

	"github.com/Kong/kuma/pkg/envoy/accesslog"
)

func defaultHandler(log logr.Logger, msg *envoy_accesslog.StreamAccessLogsMessage) (logHandler, error) {
	parts := strings.SplitN(msg.GetIdentifier().GetLogName(), ";", 2)
	if len(parts) != 2 {
		return nil, errors.Errorf("log name %q has invalid format: expected %d components separated by ';', got %d", msg.GetIdentifier().GetLogName(), 2, len(parts))
	}
	address, formatString := parts[0], parts[1]

	format, err := accesslog.ParseFormat(formatString)
	if err != nil {
		return nil, err
	}

	sender, err := defaultSender(log, address)
	if err != nil {
		return nil, err
	}
	if err := sender.Connect(); err != nil {
		return nil, err
	}

	return &handler{
		format: format,
		sender: sender,
	}, nil
}

func defaultSender(log logr.Logger, address string) (logSender, error) {
	return &sender{
		log:     log,
		address: address,
	}, nil
}
