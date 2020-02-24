package bootstrap

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/Kong/kuma/pkg/xds/topology"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	bootstrap_config "github.com/Kong/kuma/pkg/config/xds/bootstrap"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
	"github.com/Kong/kuma/pkg/xds/bootstrap/types"
	envoy_bootstrap "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v2"
)

type BootstrapGenerator interface {
	Generate(ctx context.Context, request types.BootstrapRequest) (proto.Message, error)
}

func NewDefaultBootstrapGenerator(
	resManager core_manager.ResourceManager,
	config *bootstrap_config.BootstrapParamsConfig) BootstrapGenerator {
	return &bootstrapGenerator{
		resManager: resManager,
		config:     config,
	}
}

type bootstrapGenerator struct {
	resManager core_manager.ResourceManager
	config     *bootstrap_config.BootstrapParamsConfig
}

func (b *bootstrapGenerator) Generate(ctx context.Context, request types.BootstrapRequest) (proto.Message, error) {
	proxyId, err := core_xds.BuildProxyId(request.Mesh, request.Name)
	if err != nil {
		return nil, err
	}
	dataplane, err := b.fetchDataplane(ctx, proxyId)
	if err != nil {
		return nil, err
	}
	bootstrapCfg, err := b.generateFor(*proxyId, dataplane, request)
	if err != nil {
		return nil, err
	}
	tracingBackend, err := b.fetchTracingBackend(ctx, dataplane)
	if err != nil {
		return nil, err
	}
	if err := AddTracingConfig(bootstrapCfg, tracingBackend); err != nil {
		return nil, err
	}
	return bootstrapCfg, nil
}

func (b *bootstrapGenerator) generateFor(proxyId core_xds.ProxyId, dataplane *core_mesh.DataplaneResource, request types.BootstrapRequest) (*envoy_bootstrap.Bootstrap, error) {
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
	return b.configForParameters(params)
}

func (b *bootstrapGenerator) fetchDataplane(ctx context.Context, proxyId *core_xds.ProxyId) (*core_mesh.DataplaneResource, error) {
	res := core_mesh.DataplaneResource{}
	if err := b.resManager.Get(ctx, &res, core_store.GetBy(proxyId.ToResourceKey())); err != nil {
		return nil, err
	}
	return &res, nil
}

func (b *bootstrapGenerator) fetchTracingBackend(ctx context.Context, dataplane *core_mesh.DataplaneResource) (*mesh_proto.TracingBackend, error) {
	mesh := core_mesh.MeshResource{}
	if err := b.resManager.Get(context.Background(), &mesh, core_store.GetByKey(dataplane.GetMeta().GetMesh(), dataplane.GetMeta().GetMesh())); err != nil {
		return nil, err
	}

	trafficTrace, err := topology.GetTrafficTrace(ctx, dataplane, b.resManager)
	if err != nil {
		return nil, err
	}
	if trafficTrace == nil {
		return nil, nil
	}

	return mesh.GetTracingBackend(trafficTrace.Spec.GetConf().GetBackend()), nil
}

func (b *bootstrapGenerator) configForParameters(params configParameters) (*envoy_bootstrap.Bootstrap, error) {
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
