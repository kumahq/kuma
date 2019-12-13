package envoy

import (
	"fmt"
	"strings"

	"github.com/Kong/kuma/api/mesh/v1alpha1"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v2"
	filter_accesslog "github.com/envoyproxy/go-control-plane/envoy/config/filter/accesslog/v2"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"
)

const AccessLogDefaultFormat = "[%START_TIME%] %KUMA_SOURCE_ADDRESS%(%KUMA_SOURCE_SERVICE%)->%UPSTREAM_HOST%(%KUMA_DESTINATION_SERVICE%) took %DURATION%ms, sent %BYTES_SENT% bytes, received: %BYTES_RECEIVED% bytes\n"

const AccessLogSink = "access_log_sink"

func convertLoggingBackend(sourceService string, destinationService string, backend *v1alpha1.LoggingBackend, proxy *core_xds.Proxy) (*filter_accesslog.AccessLog, error) {
	if backend == nil {
		return nil, nil
	}
	format := AccessLogDefaultFormat
	if backend.Format != "" {
		format = backend.Format
	}
	iface, _ := proxy.Dataplane.Spec.Networking.GetInboundInterface(sourceService)
	sourceAddress := ""
	if iface != nil {
		sourceAddress = iface.DataplaneIP
	}
	format = strings.ReplaceAll(format, "%KUMA_SOURCE_ADDRESS%", fmt.Sprintf("%s:0", sourceAddress))
	format = strings.ReplaceAll(format, "%KUMA_SOURCE_SERVICE%", sourceService)
	format = strings.ReplaceAll(format, "%KUMA_DESTINATION_SERVICE%", destinationService)

	if file, ok := backend.GetType().(*v1alpha1.LoggingBackend_File_); ok {
		return fileAccessLog(format, file)
	} else if tcp, ok := backend.GetType().(*v1alpha1.LoggingBackend_Tcp_); ok {
		return tcpAccessLog(format, tcp)
	} else {
		return nil, errors.Errorf("could not convert LoggingBackend of type %T to AccessLog", backend.GetType())
	}
}

func tcpAccessLog(format string, tcp *v1alpha1.LoggingBackend_Tcp_) (*filter_accesslog.AccessLog, error) {
	fileAccessLog := &accesslog.HttpGrpcAccessLogConfig{
		CommonConfig: &accesslog.CommonGrpcAccessLogConfig{
			LogName: fmt.Sprintf("%s;%s", tcp.Tcp.Address, format),
			GrpcService: &envoy_core.GrpcService{
				TargetSpecifier: &envoy_core.GrpcService_EnvoyGrpc_{
					EnvoyGrpc: &envoy_core.GrpcService_EnvoyGrpc{
						ClusterName: AccessLogSink,
					},
				},
			},
		},
	}
	marshalled, err := ptypes.MarshalAny(fileAccessLog)
	if err != nil {
		return nil, errors.Wrap(err, "could not marshall FileAccessLog")
	}
	return &filter_accesslog.AccessLog{
		Name: wellknown.HTTPGRPCAccessLog,
		ConfigType: &filter_accesslog.AccessLog_TypedConfig{
			TypedConfig: marshalled,
		},
	}, nil
}

func fileAccessLog(format string, file *v1alpha1.LoggingBackend_File_) (*filter_accesslog.AccessLog, error) {
	fileAccessLog := &accesslog.FileAccessLog{
		AccessLogFormat: &accesslog.FileAccessLog_Format{
			Format: format,
		},
		Path: file.File.Path,
	}
	marshalled, err := ptypes.MarshalAny(fileAccessLog)
	if err != nil {
		return nil, errors.Wrap(err, "could not marshall FileAccessLog")
	}
	return &filter_accesslog.AccessLog{
		Name: wellknown.FileAccessLog,
		ConfigType: &filter_accesslog.AccessLog_TypedConfig{
			TypedConfig: marshalled,
		},
	}, nil
}
