package envoy

import (
	"github.com/Kong/kuma/api/mesh/v1alpha1"
	accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v2"
	filter_accesslog "github.com/envoyproxy/go-control-plane/envoy/config/filter/accesslog/v2"
	"github.com/envoyproxy/go-control-plane/pkg/util"
	"github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
)

const AccessLogDefaultFormat = "[%START_TIME%] %DOWNSTREAM_REMOTE_ADDRESS%->%UPSTREAM_HOST%(%UPSTREAM_CLUSTER%) took %DURATION%ms, sent %BYTES_SENT% bytes, received: %BYTES_RECEIVED% bytes\n"

func convertLoggingBackends(backends []*v1alpha1.LoggingBackend) ([]*filter_accesslog.AccessLog, error) {
	var result []*filter_accesslog.AccessLog
	for _, backend := range backends {
		log, err := convertLoggingBackend(backend)
		if err != nil {
			return nil, err
		}
		result = append(result, log)
	}
	return result, nil
}

func convertLoggingBackend(backend *v1alpha1.LoggingBackend) (*filter_accesslog.AccessLog, error) {
	format := AccessLogDefaultFormat
	if backend.Format != "" {
		format = backend.Format
	}
	if file, ok := backend.GetType().(*v1alpha1.LoggingBackend_File_); ok {
		return fileAccessLog(format, file)
	} else {
		return nil, errors.Errorf("could not convert LoggingBackend of type %T to AccessLog", backend.GetType())
	}
}

func fileAccessLog(format string, file *v1alpha1.LoggingBackend_File_) (*filter_accesslog.AccessLog, error) {
	fileAccessLog := &accesslog.FileAccessLog{
		AccessLogFormat: &accesslog.FileAccessLog_Format{
			Format: format,
		},
		Path: file.File.Path,
	}
	marshalled, err := types.MarshalAny(fileAccessLog)
	if err != nil {
		return nil, errors.Wrap(err, "could not marshall FileAccessLog")
	}
	return &filter_accesslog.AccessLog{
		Name: util.FileAccessLog,
		ConfigType: &filter_accesslog.AccessLog_TypedConfig{
			TypedConfig: marshalled,
		},
	}, nil
}
