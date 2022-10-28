package v1alpha1

import (
	"net"

	envoy_accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	access_loggers_file "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/file/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_tcp "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/structpb"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	accesslog "github.com/kumahq/kuma/pkg/envoy/accesslog/v3"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	listeners_v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
)

const (
	defaultNetworkAccessLogFormat = `[%START_TIME%] %RESPONSE_FLAGS% %KUMA_MESH% %KUMA_SOURCE_ADDRESS_WITHOUT_PORT%(%KUMA_SOURCE_SERVICE%)->%UPSTREAM_HOST%(%KUMA_DESTINATION_SERVICE%) took %DURATION%ms, sent %BYTES_SENT% bytes, received: %BYTES_RECEIVED% bytes`
	defaultHttpAccessLogFormat    = `[%START_TIME%] %KUMA_MESH% "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% %RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)% "%REQ(X-FORWARDED-FOR)%" "%REQ(USER-AGENT)%" "%REQ(X-B3-TRACEID?X-DATADOG-TRACEID)%" "%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%" "%KUMA_SOURCE_SERVICE%" "%KUMA_DESTINATION_SERVICE%" "%KUMA_SOURCE_ADDRESS_WITHOUT_PORT%" "%UPSTREAM_HOST%"`
)

type Configurer struct {
	Mesh               string
	TrafficDirection   envoy.TrafficDirection
	SourceService      string
	DestinationService string
	Backend            api.Backend
	Dataplane          *core_mesh.DataplaneResource
}

func (c *Configurer) interpolateKumaVariables(formatString string) (*accesslog.AccessLogFormat, error) {
	envoyFormat, err := accesslog.ParseFormat(formatString)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid access log format string: %s", formatString)
	}

	variables := accesslog.InterpolationVariables{
		accesslog.CMD_KUMA_SOURCE_ADDRESS:              net.JoinHostPort(c.Dataplane.GetIP(), "0"), // deprecated variable
		accesslog.CMD_KUMA_SOURCE_ADDRESS_WITHOUT_PORT: c.Dataplane.GetIP(),                        // replacement variable
		accesslog.CMD_KUMA_SOURCE_SERVICE:              c.SourceService,
		accesslog.CMD_KUMA_DESTINATION_SERVICE:         c.DestinationService,
		accesslog.CMD_KUMA_MESH:                        c.Mesh,
		accesslog.CMD_KUMA_TRAFFIC_DIRECTION:           string(c.TrafficDirection),
	}

	envoyFormat, err = envoyFormat.Interpolate(variables)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to interpolate access log format string with Kuma-specific variables: %s", formatString)
	}

	return envoyFormat, nil
}

func (c *Configurer) envoyAccessLog(defaultFormat string) (*envoy_accesslog.AccessLog, error) {
	var format *api.Format
	var logPath string

	if f := c.Backend.File; f != nil {
		format = f.Format
		logPath = f.Path
	} else if f := c.Backend.Tcp; f != nil {
		format = f.Format
		logPath = envoy.AccessLogSocketName(c.Dataplane.Meta.GetName(), c.Mesh)
	}

	var substitutionFormatString envoy_core.SubstitutionFormatString

	if len(format.Json) == 0 {
		formatString := format.Plain
		if formatString == "" {
			formatString = defaultFormat
		}
		formatString += "\n"

		envoyFormat, err := c.interpolateKumaVariables(formatString)
		if err != nil {
			return nil, err
		}

		if file := c.Backend.File; file != nil {
			substitutionFormatString = envoy_core.SubstitutionFormatString{
				Format: &envoy_core.SubstitutionFormatString_TextFormatSource{
					TextFormatSource: &envoy_core.DataSource{
						Specifier: &envoy_core.DataSource_InlineString{
							InlineString: envoyFormat.String(),
						},
					},
				},
			}
		} else if tcp := c.Backend.Tcp; tcp != nil {
			substitutionFormatString = envoy_core.SubstitutionFormatString{
				Format: &envoy_core.SubstitutionFormatString_JsonFormat{
					JsonFormat: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"address": structpb.NewStringValue(tcp.Address),
							"message": structpb.NewStringValue(envoyFormat.String()),
						},
					},
				},
			}
		}
	} else {
		fields := map[string]*structpb.Value{}
		for _, kv := range format.Json {
			interpolated, err := c.interpolateKumaVariables(kv.Value)
			if err != nil {
				return nil, err
			}

			fields[kv.Key] = structpb.NewStringValue(interpolated.String())
		}

		var jsonFormat structpb.Struct
		if file := c.Backend.File; file != nil {
			jsonFormat = structpb.Struct{
				Fields: fields,
			}
		} else if tcp := c.Backend.Tcp; tcp != nil {
			jsonFormat = structpb.Struct{
				Fields: map[string]*structpb.Value{
					"address": structpb.NewStringValue(tcp.Address),
					"message": structpb.NewStructValue(&structpb.Struct{
						Fields: fields,
					}),
				},
			}
		}
		substitutionFormatString = envoy_core.SubstitutionFormatString{
			Format: &envoy_core.SubstitutionFormatString_JsonFormat{
				JsonFormat: &jsonFormat,
			},
		}
	}

	return fileAccessLog(&substitutionFormatString, logPath)
}

func fileAccessLog(logFormat *envoy_core.SubstitutionFormatString, path string) (*envoy_accesslog.AccessLog, error) {
	fileAccessLog := &access_loggers_file.FileAccessLog{
		AccessLogFormat: &access_loggers_file.FileAccessLog_LogFormat{
			LogFormat: logFormat,
		},
		Path: path,
	}

	marshaled, err := util_proto.MarshalAnyDeterministic(fileAccessLog)
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

func (c *Configurer) Configure(filterChain *envoy_listener.FilterChain) error {
	httpAccessLog := func(hcm *envoy_hcm.HttpConnectionManager) error {
		accessLog, err := c.envoyAccessLog(defaultHttpAccessLogFormat)
		if err != nil {
			return err
		}
		hcm.AccessLog = append(hcm.AccessLog, accessLog)
		return nil
	}
	tcpAccessLog := func(tcpProxy *envoy_tcp.TcpProxy) error {
		accessLog, err := c.envoyAccessLog(defaultNetworkAccessLogFormat)
		if err != nil {
			return err
		}

		tcpProxy.AccessLog = append(tcpProxy.AccessLog, accessLog)
		return nil
	}

	if err := listeners_v3.UpdateHTTPConnectionManager(filterChain, httpAccessLog); err != nil && !errors.Is(err, &listeners_v3.UnexpectedFilterConfigTypeError{}) {
		return err
	}
	if err := listeners_v3.UpdateTCPProxy(filterChain, tcpAccessLog); err != nil && !errors.Is(err, &listeners_v3.UnexpectedFilterConfigTypeError{}) {
		return err
	}

	return nil
}
