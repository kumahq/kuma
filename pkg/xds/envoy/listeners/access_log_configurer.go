package listeners

import (
	"fmt"
	"net"

	"github.com/pkg/errors"

	"github.com/golang/protobuf/ptypes"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v2"
	filter_accesslog "github.com/envoyproxy/go-control-plane/envoy/config/filter/accesslog/v2"
	envoy_wellknown "github.com/envoyproxy/go-control-plane/pkg/wellknown"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
	"github.com/Kong/kuma/pkg/envoy/accesslog"
)

const accessLogSink = "access_log_sink"

type AccessLogConfigurer struct {
	mesh               string
	trafficDirection   string
	sourceService      string
	destinationService string
	backend            *mesh_proto.LoggingBackend
	proxy              *core_xds.Proxy
}

func convertLoggingBackend(mesh string, trafficDirection string, sourceService string, destinationService string, backend *mesh_proto.LoggingBackend, proxy *core_xds.Proxy, defaultFormat string) (*filter_accesslog.AccessLog, error) {
	if backend == nil {
		return nil, nil
	}
	formatString := defaultFormat
	if backend.Format != "" {
		formatString = backend.Format
	}
	format, err := accesslog.ParseFormat(formatString)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid access log format string: %s", formatString)
	}

	variables := accesslog.InterpolationVariables{
		accesslog.CMD_KUMA_SOURCE_ADDRESS:              net.JoinHostPort(proxy.Dataplane.GetIP(), "0"), // deprecated variable
		accesslog.CMD_KUMA_SOURCE_ADDRESS_WITHOUT_PORT: proxy.Dataplane.GetIP(),                        // replacement variable
		accesslog.CMD_KUMA_SOURCE_SERVICE:              sourceService,
		accesslog.CMD_KUMA_DESTINATION_SERVICE:         destinationService,
		accesslog.CMD_KUMA_MESH:                        mesh,
		accesslog.CMD_KUMA_TRAFFIC_DIRECTION:           trafficDirection,
	}

	format, err = format.Interpolate(variables)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to interpolate access log format string with Kuma-specific variables: %s", formatString)
	}

	if file, ok := backend.GetType().(*mesh_proto.LoggingBackend_File_); ok {
		return fileAccessLog(format, file)
	} else if tcp, ok := backend.GetType().(*mesh_proto.LoggingBackend_Tcp_); ok {
		return tcpAccessLog(format, tcp)
	} else {
		return nil, errors.Errorf("could not convert LoggingBackend of type %T to AccessLog", backend.GetType())
	}
}

func tcpAccessLog(format *accesslog.AccessLogFormat, tcp *mesh_proto.LoggingBackend_Tcp_) (*filter_accesslog.AccessLog, error) {
	httpGrpcAccessLog := &envoy_accesslog.HttpGrpcAccessLogConfig{
		CommonConfig: &envoy_accesslog.CommonGrpcAccessLogConfig{
			LogName: fmt.Sprintf("%s;%s", tcp.Tcp.Address, format.String()),
			GrpcService: &envoy_core.GrpcService{
				TargetSpecifier: &envoy_core.GrpcService_EnvoyGrpc_{
					EnvoyGrpc: &envoy_core.GrpcService_EnvoyGrpc{
						ClusterName: accessLogSink,
					},
				},
			},
		},
	}
	if err := format.ConfigureHttpLog(httpGrpcAccessLog); err != nil {
		return nil, errors.Wrapf(err, "failed to configure %T according to the format string: %s", httpGrpcAccessLog, format)
	}
	marshalled, err := ptypes.MarshalAny(httpGrpcAccessLog)
	if err != nil {
		return nil, errors.Wrapf(err, "could not marshall %T", httpGrpcAccessLog)
	}
	return &filter_accesslog.AccessLog{
		Name: envoy_wellknown.HTTPGRPCAccessLog,
		ConfigType: &filter_accesslog.AccessLog_TypedConfig{
			TypedConfig: marshalled,
		},
	}, nil
}

func fileAccessLog(format *accesslog.AccessLogFormat, file *mesh_proto.LoggingBackend_File_) (*filter_accesslog.AccessLog, error) {
	fileAccessLog := &envoy_accesslog.FileAccessLog{
		AccessLogFormat: &envoy_accesslog.FileAccessLog_Format{
			Format: format.String(),
		},
		Path: file.File.Path,
	}
	marshalled, err := ptypes.MarshalAny(fileAccessLog)
	if err != nil {
		return nil, errors.Wrapf(err, "could not marshall %T", fileAccessLog)
	}
	return &filter_accesslog.AccessLog{
		Name: envoy_wellknown.FileAccessLog,
		ConfigType: &filter_accesslog.AccessLog_TypedConfig{
			TypedConfig: marshalled,
		},
	}, nil
}
