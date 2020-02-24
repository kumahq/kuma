package accesslogs

import (
	"github.com/pkg/errors"

	envoy_accesslog "github.com/envoyproxy/go-control-plane/envoy/service/accesslog/v2"

	"github.com/Kong/kuma/pkg/envoy/accesslog"
)

type handler struct {
	format *accesslog.AccessLogFormat
	sender logSender
}

func (h *handler) Handle(msg *envoy_accesslog.StreamAccessLogsMessage) error {
	switch logEntries := msg.GetLogEntries().(type) {
	case *envoy_accesslog.StreamAccessLogsMessage_HttpLogs:
		for _, httpLogEntry := range logEntries.HttpLogs.GetLogEntry() {
			record, err := h.format.FormatHttpLogEntry(httpLogEntry)
			if err != nil {
				return errors.Wrapf(err, "failed to format an HTTP log entry %v as %q", httpLogEntry, h.format)
			}
			if err := h.sender.Send(record); err != nil {
				return err
			}
		}
	case *envoy_accesslog.StreamAccessLogsMessage_TcpLogs:
		for _, tcpLogEntry := range logEntries.TcpLogs.GetLogEntry() {
			record, err := h.format.FormatTcpLogEntry(tcpLogEntry)
			if err != nil {
				return errors.Wrapf(err, "failed to format a TCP log entry %v as %q", tcpLogEntry, h.format)
			}
			if err := h.sender.Send(record); err != nil {
				return err
			}
		}
	default:
		return errors.Errorf("unknown type of log entries: %T", msg.GetLogEntries())
	}
	return nil
}

func (h *handler) Close() error {
	return h.sender.Close()
}
