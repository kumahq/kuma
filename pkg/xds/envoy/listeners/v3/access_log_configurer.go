package v3

import (
	"fmt"
	"net"
	"strings"

	envoy_accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	access_loggers_file "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/file/v3"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
)

const (
	CMD_KUMA_SOURCE_ADDRESS              = "%KUMA_SOURCE_ADDRESS%"
	CMD_KUMA_SOURCE_ADDRESS_WITHOUT_PORT = "%KUMA_SOURCE_ADDRESS_WITHOUT_PORT%"
	CMD_KUMA_SOURCE_SERVICE              = "%KUMA_SOURCE_SERVICE%"
	CMD_KUMA_DESTINATION_SERVICE         = "%KUMA_DESTINATION_SERVICE%"
	CMD_KUMA_MESH                        = "%KUMA_MESH%"
	CMD_KUMA_TRAFFIC_DIRECTION           = "%KUMA_TRAFFIC_DIRECTION%"
)

type AccessLogConfigurer struct {
	Mesh               string
	TrafficDirection   envoy.TrafficDirection
	SourceService      string
	DestinationService string
	Backend            *mesh_proto.LoggingBackend
	Proxy              *core_xds.Proxy
}

type KumaValues struct {
	SourceService      string
	SourceIP           string
	DestinationService string
	Mesh               string
	TrafficDirection   envoy.TrafficDirection
}

func InterpolateKumaValues(format string, values KumaValues) string {
	format = strings.ReplaceAll(format, CMD_KUMA_SOURCE_ADDRESS, net.JoinHostPort(values.SourceIP, "0"))
	format = strings.ReplaceAll(format, CMD_KUMA_SOURCE_ADDRESS_WITHOUT_PORT, values.SourceIP)
	format = strings.ReplaceAll(format, CMD_KUMA_SOURCE_SERVICE, values.SourceService)
	format = strings.ReplaceAll(format, CMD_KUMA_DESTINATION_SERVICE, values.DestinationService)
	format = strings.ReplaceAll(format, CMD_KUMA_MESH, values.Mesh)
	format = strings.ReplaceAll(format, CMD_KUMA_TRAFFIC_DIRECTION, string(values.TrafficDirection))
	return format
}

func convertLoggingBackend(mesh string, trafficDirection envoy.TrafficDirection, sourceService string, destinationService string, backend *mesh_proto.LoggingBackend, proxy *core_xds.Proxy, defaultFormat string) (*envoy_accesslog.AccessLog, error) {
	if backend == nil {
		return nil, nil
	}
	format := defaultFormat
	if backend.Format != "" {
		format = backend.Format
	}
	values := KumaValues{
		SourceService:      sourceService,
		SourceIP:           proxy.Dataplane.GetIP(),
		DestinationService: destinationService,
		Mesh:               mesh,
		TrafficDirection:   trafficDirection,
	}
	format = InterpolateKumaValues(format, values)
	format += "\n"

	switch backend.GetType() {
	case mesh_proto.LoggingFileType:
		cfg := mesh_proto.FileLoggingBackendConfig{}
		if err := proto.ToTyped(backend.Conf, &cfg); err != nil {
			return nil, errors.Wrap(err, "could not parse backend config")
		}
		return fileAccessLog(format, cfg.Path)
	case mesh_proto.LoggingTcpType:
		cfg := mesh_proto.TcpLoggingBackendConfig{}
		if err := proto.ToTyped(backend.Conf, &cfg); err != nil {
			return nil, errors.Wrap(err, "could not parse backend config")
		}
		accessLogSocketPath := core_xds.AccessLogSocketName(proxy.Metadata.WorkDir, proxy.Id.ToResourceKey().Name, proxy.Id.ToResourceKey().Mesh)
		return fileAccessLog(fmt.Sprintf("%s;%s", cfg.Address, format), accessLogSocketPath)
	default: // should be caught by validator
		return nil, errors.Errorf("could not convert LoggingBackend of type %T to AccessLog", backend.GetType())
	}
}

func fileAccessLog(format string, path string) (*envoy_accesslog.AccessLog, error) {
	fileAccessLog := &access_loggers_file.FileAccessLog{
		AccessLogFormat: &access_loggers_file.FileAccessLog_LogFormat{
			LogFormat: &envoy_core.SubstitutionFormatString{
				Format: &envoy_core.SubstitutionFormatString_TextFormatSource{
					TextFormatSource: &envoy_core.DataSource{
						Specifier: &envoy_core.DataSource_InlineString{
							InlineString: format,
						},
					},
				},
			},
		},
		Path: path,
	}

	marshaled, err := proto.MarshalAnyDeterministic(fileAccessLog)
	if err != nil {
		return nil, errors.Wrapf(err, "could not marshall %T", fileAccessLog)
	}
	return &envoy_accesslog.AccessLog{
		Name: "envoy.access_loggers.file",
		ConfigType: &envoy_accesslog.AccessLog_TypedConfig{
			TypedConfig: marshaled,
		},
	}, nil
}
