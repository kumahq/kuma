package accesslog

import (
	"strings"

	accesslog_config "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v2"
	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"
)

const (
	unspecifiedValue = "-"
)

type CompositeLogConfigureFormatter struct {
	Formatters []LogEntryFormatter
}

func (c *CompositeLogConfigureFormatter) ConfigureHttpLog(config *accesslog_config.HttpGrpcAccessLogConfig) error {
	for _, formatter := range c.Formatters {
		if configurer, ok := formatter.(HttpLogConfigurer); ok {
			if err := configurer.ConfigureHttpLog(config); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *CompositeLogConfigureFormatter) ConfigureTcpLog(config *accesslog_config.TcpGrpcAccessLogConfig) error {
	for _, formatter := range c.Formatters {
		if configurer, ok := formatter.(TcpLogConfigurer); ok {
			if err := configurer.ConfigureTcpLog(config); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *CompositeLogConfigureFormatter) FormatHttpLogEntry(entry *accesslog_data.HTTPAccessLogEntry) (string, error) {
	values := make([]string, len(c.Formatters))
	for i, formatter := range c.Formatters {
		value, err := formatter.FormatHttpLogEntry(entry)
		if err != nil {
			return "", err
		}
		if value == "" {
			value = unspecifiedValue
		}
		values[i] = value
	}
	return strings.Join(values, ""), nil
}

func (c *CompositeLogConfigureFormatter) FormatTcpLogEntry(entry *accesslog_data.TCPAccessLogEntry) (string, error) {
	values := make([]string, len(c.Formatters))
	for i, formatter := range c.Formatters {
		value, err := formatter.FormatTcpLogEntry(entry)
		if err != nil {
			return "", err
		}
		if value == "" {
			value = unspecifiedValue
		}
		values[i] = value
	}
	return strings.Join(values, ""), nil
}
