package bootstrap

import (
	"bytes"
	"context"
	mesh_proto "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	xds_config "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/xds"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/manager"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/xds"
	util_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/proto"
	envoy_bootstrap "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v2"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"text/template"
)

type BootstrapRequest struct {
	NodeId string `json:"nodeId"`
}

type BootstrapGenerator interface {
	Generate(ctx context.Context, request BootstrapRequest) (proto.Message, error)
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

func (b *bootstrapGenerator) Generate(ctx context.Context, request BootstrapRequest) (proto.Message, error) {
	dataplane, err := b.fetchDataplane(ctx, request.NodeId)
	if err != nil {
		return nil, err
	}

	// if dataplane has no service - fill this with placeholder. Otherwise take the first service
	service := "unknown"
	services := dataplane.Spec.Tags().Values(mesh_proto.ServiceTag)
	if len(services) > 0 {
		service = services[0]
	}
	params := configParameters{
		Id:        request.NodeId,
		Service:   service,
		AdminPort: b.config.AdminPort,
		XdsHost:   b.config.XdsHost,
		XdsPort:   b.config.XdsPort,
	}
	return b.ConfigForParameters(params)
}

func (b *bootstrapGenerator) fetchDataplane(ctx context.Context, nodeId string) (*mesh.DataplaneResource, error) {
	id, err := xds.ParseProxyIdFromString(nodeId)
	if err != nil {
		return nil, err
	}
	res := mesh.DataplaneResource{}
	if err := b.resManager.Get(ctx, &res, store.GetBy(id.ToResourceKey())); err != nil {
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
