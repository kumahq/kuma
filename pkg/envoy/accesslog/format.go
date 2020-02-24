package accesslog

import (
	"strings"

	accesslog_config "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v2"
	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"
)

const (
	unspecifiedValue = "-" // to replicate Envoy's behaviour
)

// AccessLogFormat represents the entire access log format string.
type AccessLogFormat struct {
	Fragments []AccessLogFragment
}

func (f *AccessLogFormat) FormatHttpLogEntry(entry *accesslog_data.HTTPAccessLogEntry) (string, error) {
	values := make([]string, len(f.Fragments))
	for i, fragment := range f.Fragments {
		value, err := fragment.FormatHttpLogEntry(entry)
		if err != nil {
			return "", err
		}
		if value == "" {
			value = unspecifiedValue // to replicate Envoy's behaviour
		}
		values[i] = value
	}
	return strings.Join(values, ""), nil
}

func (f *AccessLogFormat) FormatTcpLogEntry(entry *accesslog_data.TCPAccessLogEntry) (string, error) {
	values := make([]string, len(f.Fragments))
	for i, fragment := range f.Fragments {
		value, err := fragment.FormatTcpLogEntry(entry)
		if err != nil {
			return "", err
		}
		if value == "" {
			value = unspecifiedValue // to replicate Envoy's behaviour
		}
		values[i] = value
	}
	return strings.Join(values, ""), nil
}

func (f *AccessLogFormat) ConfigureHttpLog(config *accesslog_config.HttpGrpcAccessLogConfig) error {
	for _, fragment := range f.Fragments {
		if err := fragment.ConfigureHttpLog(config); err != nil {
			return err
		}
	}
	return nil
}

func (f *AccessLogFormat) ConfigureTcpLog(config *accesslog_config.TcpGrpcAccessLogConfig) error {
	for _, fragment := range f.Fragments {
		if err := fragment.ConfigureTcpLog(config); err != nil {
			return err
		}
	}
	return nil
}

func (f *AccessLogFormat) Interpolate(variables InterpolationVariables) (*AccessLogFormat, error) {
	newFragments := make([]AccessLogFragment, len(f.Fragments))
	interpolated := false
	for i, fragment := range f.Fragments {
		if interpolator, ok := fragment.(AccessLogFragmentInterpolator); ok {
			newFragment, err := interpolator.Interpolate(variables)
			if err != nil {
				return nil, err
			}
			newFragments[i] = newFragment
			interpolated = interpolated || newFragment != fragment
		} else {
			newFragments[i] = fragment
		}
	}
	if interpolated {
		return &AccessLogFormat{Fragments: newFragments}, nil
	}
	return f, nil
}

// String returns the canonical representation of this format string.
func (f *AccessLogFormat) String() string {
	fragments := make([]string, len(f.Fragments))
	for i, fragment := range f.Fragments {
		fragments[i] = fragment.String()
	}
	return strings.Join(fragments, "")
}
