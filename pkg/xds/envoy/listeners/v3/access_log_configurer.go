package v3

import (
	"fmt"
	"net"

	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/pkg/errors"

	envoy_accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	access_loggers_file "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/file/v3"
	access_loggers_grpc "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/grpc/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	accesslog "github.com/kumahq/kuma/pkg/envoy/accesslog/v3"
	"github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
)

const accessLogSink = "access_log_sink"

type AccessLogConfigurer struct {
	Mesh               string
	TrafficDirection   envoy.TrafficDirection
	SourceService      string
	DestinationService string
	Backend            *mesh_proto.LoggingBackend
	Proxy              *core_xds.Proxy
}

func convertLoggingBackend(mesh string, trafficDirection envoy.TrafficDirection, sourceService string, destinationService string, backend *mesh_proto.LoggingBackend, proxy *core_xds.Proxy, defaultFormat string) (*envoy_accesslog.AccessLog, error) {
	if backend == nil {
		return nil, nil
	}
	formatString := defaultFormat
	if backend.Format != "" {
		formatString = backend.Format
	}
	format, err := accesslog.ParseFormat(formatString + "\n")

	if err != nil {
		return nil, errors.Wrapf(err, "invalid access log format string: %s", formatString)
	}

	variables := accesslog.InterpolationVariables{
		accesslog.CMD_KUMA_SOURCE_ADDRESS:              net.JoinHostPort(proxy.Dataplane.GetIP(), "0"), // deprecated variable
		accesslog.CMD_KUMA_SOURCE_ADDRESS_WITHOUT_PORT: proxy.Dataplane.GetIP(),                        // replacement variable
		accesslog.CMD_KUMA_SOURCE_SERVICE:              sourceService,
		accesslog.CMD_KUMA_DESTINATION_SERVICE:         destinationService,
		accesslog.CMD_KUMA_MESH:                        mesh,
		accesslog.CMD_KUMA_TRAFFIC_DIRECTION:           string(trafficDirection),
	}

	format, err = format.Interpolate(variables)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to interpolate access log format string with Kuma-specific variables: %s", formatString)
	}

	switch backend.GetType() {
	case mesh_proto.LoggingFileType:
		return fileAccessLog(format, backend.Conf)
	case mesh_proto.LoggingTcpType:
		return tcpAccessLog(format, backend.Conf)
	default: // should be caught by validator
		return nil, errors.Errorf("could not convert LoggingBackend of type %T to AccessLog", backend.GetType())
	}
}

func tcpAccessLog(format *accesslog.AccessLogFormat, cfgStr *structpb.Struct) (*envoy_accesslog.AccessLog, error) {
	cfg := mesh_proto.TcpLoggingBackendConfig{}
	if err := proto.ToTyped(cfgStr, &cfg); err != nil {
		return nil, errors.Wrap(err, "could not parse backend config")
	}

	httpGrpcAccessLog := &access_loggers_grpc.HttpGrpcAccessLogConfig{
		CommonConfig: &access_loggers_grpc.CommonGrpcAccessLogConfig{
			LogName:             fmt.Sprintf("%s;%s", cfg.Address, format.String()),
			TransportApiVersion: envoy_core.ApiVersion_V3,
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
	marshalled, err := proto.MarshalAnyDeterministic(httpGrpcAccessLog)
	if err != nil {
		return nil, errors.Wrapf(err, "could not marshall %T", httpGrpcAccessLog)
	}
	return &envoy_accesslog.AccessLog{
		Name: "envoy.access_loggers.http_grpc",
		ConfigType: &envoy_accesslog.AccessLog_TypedConfig{
			TypedConfig: marshalled,
		},
	}, nil
}

func fileAccessLog(format *accesslog.AccessLogFormat, cfgStr *structpb.Struct) (*envoy_accesslog.AccessLog, error) {
	cfg := mesh_proto.FileLoggingBackendConfig{}
	if err := proto.ToTyped(cfgStr, &cfg); err != nil {
		return nil, errors.Wrap(err, "could not parse backend config")
	}

	fileAccessLog := &access_loggers_file.FileAccessLog{
		AccessLogFormat: &access_loggers_file.FileAccessLog_LogFormat{
			LogFormat: &envoy_core.SubstitutionFormatString{
				Format: &envoy_core.SubstitutionFormatString_TextFormatSource{
					TextFormatSource: &envoy_core.DataSource{
						Specifier: &envoy_core.DataSource_InlineString{
							InlineString: format.String(),
						},
					},
				},
			},
		},
		Path: cfg.Path,
	}
	marshalled, err := proto.MarshalAnyDeterministic(fileAccessLog)
	if err != nil {
		return nil, errors.Wrapf(err, "could not marshall %T", fileAccessLog)
	}
	return &envoy_accesslog.AccessLog{
		Name: "envoy.access_loggers.file",
		ConfigType: &envoy_accesslog.AccessLog_TypedConfig{
			TypedConfig: marshalled,
		},
	}, nil
}
