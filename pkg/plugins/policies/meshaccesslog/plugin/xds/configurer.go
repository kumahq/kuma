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
	"github.com/kumahq/kuma/pkg/core/validators"
	accesslog "github.com/kumahq/kuma/pkg/envoy/accesslog/v3"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
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
	switch {
	case c.Backend.Tcp != nil:
		return c.tcpBackend(c.Backend.Tcp, defaultFormat)
	case c.Backend.File != nil:
		return c.fileBackend(c.Backend.File, defaultFormat)
	default:
		return nil, errors.New(validators.MustHaveOnlyOne("backend", "tcp", "file"))
	}
}

func (c *Configurer) tcpBackend(backend *api.TCPBackend, defaultFormat string) (*envoy_accesslog.AccessLog, error) {
	var sfs *envoy_core.SubstitutionFormatString

	switch {
	case backend.Format == nil:
		envoyFormat, err := c.interpolateKumaVariables(newLine(defaultFormat))
		if err != nil {
			return nil, err
		}
		sfs = c.sfsJSON(map[string]*structpb.Value{
			"address": structpb.NewStringValue(backend.Address),
			"message": structpb.NewStringValue(envoyFormat.String()),
		}, false)
	case backend.Format.Plain != nil:
		envoyFormat, err := c.interpolateKumaVariables(newLine(*backend.Format.Plain))
		if err != nil {
			return nil, err
		}
		sfs = c.sfsJSON(map[string]*structpb.Value{
			"address": structpb.NewStringValue(backend.Address),
			"message": structpb.NewStringValue(envoyFormat.String()),
		}, pointer.Deref(backend.Format.OmitEmptyValues))
	case backend.Format.Json != nil:
		if fields, err := c.jsonToFields(*backend.Format.Json); err != nil {
			return nil, err
		} else {
			sfs = c.sfsJSON(map[string]*structpb.Value{
				"address": structpb.NewStringValue(backend.Address),
				"message": structpb.NewStructValue(&structpb.Struct{Fields: fields}),
			}, pointer.Deref(backend.Format.OmitEmptyValues))
		}
	default:
		return nil, errors.New(validators.MustHaveOnlyOne("format", "plain", "json"))
	}

	return fileAccessLog(sfs, envoy.AccessLogSocketName(c.Dataplane.Meta.GetName(), c.Mesh))
}

func (c *Configurer) fileBackend(backend *api.FileBackend, defaultFormat string) (*envoy_accesslog.AccessLog, error) {
	var sfs *envoy_core.SubstitutionFormatString

	switch {
	case backend.Format == nil:
		if plain, err := c.sfsPlain(newLine(defaultFormat), false); err != nil {
			return nil, err
		} else {
			sfs = plain
		}
	case backend.Format.Plain != nil:
		if plain, err := c.sfsPlain(newLine(*backend.Format.Plain), pointer.Deref(backend.Format.OmitEmptyValues)); err != nil {
			return nil, err
		} else {
			sfs = plain
		}
	case backend.Format.Json != nil:
		if fields, err := c.jsonToFields(*backend.Format.Json); err != nil {
			return nil, err
		} else {
			sfs = c.sfsJSON(fields, pointer.Deref(backend.Format.OmitEmptyValues))
		}
	default:
		return nil, errors.New(validators.MustHaveOnlyOne("format", "plain", "json"))
	}

	return fileAccessLog(sfs, backend.Path)
}

func newLine(s string) string {
	return s + "\n"
}

func (c *Configurer) sfsPlain(plain string, omitEmpty bool) (*envoy_core.SubstitutionFormatString, error) {
	envoyFormat, err := c.interpolateKumaVariables(plain)
	if err != nil {
		return nil, err
	}
	return &envoy_core.SubstitutionFormatString{
		Format: &envoy_core.SubstitutionFormatString_TextFormatSource{
			TextFormatSource: &envoy_core.DataSource{
				Specifier: &envoy_core.DataSource_InlineString{
					InlineString: envoyFormat.String(),
				},
			},
		},
		OmitEmptyValues: omitEmpty,
	}, nil
}

func (c *Configurer) sfsJSON(fields map[string]*structpb.Value, omitEmpty bool) *envoy_core.SubstitutionFormatString {
	return &envoy_core.SubstitutionFormatString{
		Format: &envoy_core.SubstitutionFormatString_JsonFormat{
			JsonFormat: &structpb.Struct{
				Fields: fields,
			},
		},
		OmitEmptyValues: omitEmpty,
	}
}

func (c *Configurer) jsonToFields(jsonValues []api.JsonValue) (map[string]*structpb.Value, error) {
	fields := map[string]*structpb.Value{}
	for _, kv := range jsonValues {
		interpolated, err := c.interpolateKumaVariables(kv.Value)
		if err != nil {
			return nil, err
		}

		fields[kv.Key] = structpb.NewStringValue(interpolated.String())
	}
	return fields, nil
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
