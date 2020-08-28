package bootstrap

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"text/template"

	"github.com/ghodss/yaml"

	rest_types "github.com/kumahq/kuma/pkg/core/resources/model/rest"

	"github.com/kumahq/kuma/pkg/config/core"

	envoy_bootstrap "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v2"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	bootstrap_config "github.com/kumahq/kuma/pkg/config/xds/bootstrap"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/bootstrap/types"
)

type BootstrapGenerator interface {
	Generate(ctx context.Context, request types.BootstrapRequest) (proto.Message, error)
}

func NewDefaultBootstrapGenerator(
	resManager core_manager.ResourceManager,
	config *bootstrap_config.BootstrapParamsConfig,
	cacertFile string,
	environmentType core.EnvironmentType) BootstrapGenerator {
	return &bootstrapGenerator{
		resManager:      resManager,
		config:          config,
		xdsCertFile:     cacertFile,
		environmentType: environmentType,
	}
}

type bootstrapGenerator struct {
	resManager      core_manager.ResourceManager
	config          *bootstrap_config.BootstrapParamsConfig
	xdsCertFile     string
	environmentType core.EnvironmentType
}

func (b *bootstrapGenerator) Generate(ctx context.Context, request types.BootstrapRequest) (proto.Message, error) {
	proxyId, err := core_xds.BuildProxyId(request.Mesh, request.Name)
	if err != nil {
		return nil, err
	}
	var dataplane *core_mesh.DataplaneResource
	switch b.environmentType {
	case core.KubernetesEnvironment:
		if dataplane, err = b.fetchDataplane(ctx, proxyId); err != nil {
			return nil, err
		}
	case core.UniversalEnvironment:
		dpYAML, err := base64.StdEncoding.DecodeString(request.DataplaneResource)
		if err != nil {
			return nil, err
		}
		if dataplane, err = core_mesh.ParseDataplaneYAML(dpYAML); err != nil {
			return nil, err
		}
	default:
		return nil, errors.Errorf("unknown environment type %s", b.environmentType)
	}

	bootstrapCfg, err := b.generateFor(*proxyId, dataplane, request)
	if err != nil {
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

	if err := b.verifyAdminPort(adminPort, dataplane); err != nil {
		return nil, err
	}

	var certBytes string = ""
	if b.xdsCertFile != "" {
		cert, err := ioutil.ReadFile(b.xdsCertFile)
		if err != nil {
			return nil, err
		}
		certBytes = base64.StdEncoding.EncodeToString(cert)
	}
	dpYAML, err := yaml.Marshal(rest_types.From.Resource(dataplane))
	if err != nil {
		return nil, err
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
		DataplaneResource:  base64.StdEncoding.EncodeToString(dpYAML),
		CertBytes:          certBytes,
	}
	log.WithValues("params", params).Info("Generating bootstrap config")
	return b.configForParameters(params)
}

func (b *bootstrapGenerator) verifyAdminPort(adminPort uint32, dataplane *core_mesh.DataplaneResource) error {
	//The admin port in kuma-dp is always bound to 127.0.0.1
	if dataplane.UsesInboundInterface(core_mesh.IPv4Loopback, adminPort) {
		return errors.Errorf("Resource precondition failed: Port %d requested as both admin and inbound port.", adminPort)
	}

	if dataplane.UsesOutboundInterface(core_mesh.IPv4Loopback, adminPort) {
		return errors.Errorf("Resource precondition failed: Port %d requested as both admin and outbound port.", adminPort)
	}
	return nil
}

func (b *bootstrapGenerator) fetchDataplane(ctx context.Context, proxyId *core_xds.ProxyId) (*core_mesh.DataplaneResource, error) {
	res := core_mesh.DataplaneResource{}
	if err := b.resManager.Get(ctx, &res, core_store.GetBy(proxyId.ToResourceKey())); err != nil {
		return nil, err
	}
	return &res, nil
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
