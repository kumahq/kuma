package bootstrap

import (
	"bytes"
	"context"
	"github.com/Kong/kuma/pkg/xds/bootstrap/rest"
	"text/template"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	xds_config "github.com/Kong/kuma/pkg/config/xds"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/core/xds"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
	envoy_bootstrap "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v2"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
)

type BootstrapGenerator interface {
	Generate(ctx context.Context, request rest.BootstrapRequest) (proto.Message, error)
}

func NewDefaultBootstrapGenerator(
	resManager manager.ResourceManager,
	config *xds_config.BootstrapParamsConfig) BootstrapGenerator {
	return &bootstrapGenerator{
		resManager: resManager,
		config:     config,
	}
}

type bootstrapGenerator struct {
	resManager manager.ResourceManager
	config     *xds_config.BootstrapParamsConfig
}

func (b *bootstrapGenerator) Generate(ctx context.Context, request rest.BootstrapRequest) (proto.Message, error) {
	proxyId, err := xds.BuildProxyId(request.Mesh, request.Name)
	if err != nil {
		return nil, err
	}
	dataplane, err := b.fetchDataplane(ctx, proxyId)
	if err != nil {
		return nil, err
	}

	// if dataplane has no service - fill this with placeholder. Otherwise take the first service
	service := "unknown"
	services := dataplane.Spec.Tags().Values(mesh_proto.ServiceTag)
	if len(services) > 0 {
		service = services[0]
	}

	adminPort := b.config.AdminPort
	if request.AdminPort != 0 {
		adminPort = request.AdminPort
	}
	params := configParameters{
		Id:        proxyId.String(),
		Service:   service,
		AdminPort: adminPort,
		XdsHost:   b.config.XdsHost,
		XdsPort:   b.config.XdsPort,
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
