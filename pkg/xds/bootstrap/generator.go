package bootstrap

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	bootstrap_config "github.com/Kong/kuma/pkg/config/xds/bootstrap"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/core/xds"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
	"github.com/Kong/kuma/pkg/xds/bootstrap/types"
	envoy_bootstrap "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v2"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
)

type BootstrapGenerator interface {
	Generate(ctx context.Context, request types.BootstrapRequest) (proto.Message, error)
}

func NewDefaultBootstrapGenerator(
	resManager manager.ResourceManager,
	config *bootstrap_config.BootstrapParamsConfig) BootstrapGenerator {
	return &bootstrapGenerator{
		resManager: resManager,
		config:     config,
	}
}

type bootstrapGenerator struct {
	resManager manager.ResourceManager
	config     *bootstrap_config.BootstrapParamsConfig
}

func (b *bootstrapGenerator) Generate(ctx context.Context, request types.BootstrapRequest) (proto.Message, error) {
	proxyId, err := xds.BuildProxyId(request.Mesh, request.Name)
	if err != nil {
		return nil, err
	}
	dataplane, err := b.fetchDataplane(ctx, proxyId)
	if err != nil {
		return nil, err
	}
	return b.GenerateFor(*proxyId, dataplane, request)
}

func (b *bootstrapGenerator) GenerateFor(proxyId xds.ProxyId, dataplane *mesh.DataplaneResource, request types.BootstrapRequest) (proto.Message, error) {
	// if dataplane has no service - fill this with placeholder. Otherwise take the first service
	service := dataplane.Spec.GetIdentifyingService()

	adminPort := b.config.AdminPort
	if request.AdminPort != 0 {
		adminPort = request.AdminPort
	}
	accessLogPipe := fmt.Sprintf("/tmp/kuma-access-logs-%s-%s.sock", request.Name, request.Mesh)
	params := configParameters{
		Id:                 proxyId.String(),
		Service:            service,
		AdminAddress:       b.config.AdminAddress,
		AdminPort:          adminPort,
		AdminAccessLogPath: b.config.AdminAccessLogPath,
		XdsHost:            b.config.XdsHost,
		XdsPort:            b.config.XdsPort,
		XdsConnectTimeout:  b.config.XdsConnectTimeout,
		AccessLogPipe:      accessLogPipe,
		DataplaneTokenPath: request.DataplaneTokenPath,
	}
	log.WithValues("params", params).Info("Generating bootstrap config")
	return b.ConfigForParameters(params)
}

func (b *bootstrapGenerator) fetchDataplane(ctx context.Context, proxyId *xds.ProxyId) (*mesh.DataplaneResource, error) {
	res := mesh.DataplaneResource{}
	if err := b.resManager.Get(ctx, &res, store.GetBy(proxyId.ToResourceKey())); err != nil {
		return nil, err
	}
	return &res, nil
}

func (b *bootstrapGenerator) ConfigForParameters(params configParameters) (proto.Message, error) {
	tmpl, err := template.New("bootstrap").Parse(configTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse config template")
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, params); err != nil {
		return nil, errors.Wrap(err, "failed to render config template")
	}
	config := &envoy_bootstrap.Bootstrap{}
	if err := util_proto.FromYAML(buf.Bytes(), config); err != nil {
		return nil, errors.Wrap(err, "failed to parse bootstrap config")
	}
	if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "Envoy bootstrap config is not valid")
	}
	return config, nil
}
