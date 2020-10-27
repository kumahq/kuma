package bootstrap

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"text/template"

	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/validators"

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
	_ "github.com/kumahq/kuma/pkg/xds/envoy" // import Envoy protobuf definitions so (un)marshalling Envoy protobuf works in tests (normally it is imported in root.go)
)

type BootstrapGenerator interface {
	Generate(ctx context.Context, request types.BootstrapRequest) (proto.Message, error)
}

func NewDefaultBootstrapGenerator(
	resManager core_manager.ResourceManager,
	config *bootstrap_config.BootstrapParamsConfig,
	cacertFile string,
	dpAuthEnabled bool,
	zone string,
) BootstrapGenerator {
	return &bootstrapGenerator{
		resManager:    resManager,
		config:        config,
		xdsCertFile:   cacertFile,
		dpAuthEnabled: dpAuthEnabled,
		zone:          zone,
	}
}

type bootstrapGenerator struct {
	resManager    core_manager.ResourceManager
	config        *bootstrap_config.BootstrapParamsConfig
	dpAuthEnabled bool
	xdsCertFile   string
	zone          string
}

func (b *bootstrapGenerator) Generate(ctx context.Context, request types.BootstrapRequest) (proto.Message, error) {
	if err := b.validateRequest(request); err != nil {
		return nil, err
	}

	proxyId, err := core_xds.BuildProxyId(request.Mesh, request.Name)
	if err != nil {
		return nil, err
	}

	dataplane, err := b.dataplaneFor(ctx, request, proxyId)
	if err != nil {
		return nil, err
	}

	bootstrapCfg, err := b.generateFor(*proxyId, dataplane, request)
	if err != nil {
		return nil, err
	}
	return bootstrapCfg, nil
}

var DpTokenRequired = errors.New("Dataplane Token is required. Generate token using 'kumactl generate dataplane-token > /path/file' and provide it via --dataplane-token-file=/path/file argument to Kuma DP")

func (b *bootstrapGenerator) validateRequest(request types.BootstrapRequest) error {
	if b.dpAuthEnabled && request.DataplaneTokenPath == "" {
		return DpTokenRequired
	}
	return nil
}

// dataplaneFor returns dataplane for two flows
// 1) Dataplane is passed to kuma-dp run, in this case we just read DP from the BootstrapRequest
// 2) Dataplane is created before kuma-dp run, in this case we access storage to fetch it (ex. Kubernetes)
func (b *bootstrapGenerator) dataplaneFor(ctx context.Context, request types.BootstrapRequest, proxyId *core_xds.ProxyId) (*core_mesh.DataplaneResource, error) {
	if request.DataplaneResource != "" {
		res, err := rest.UnmarshallToCore([]byte(request.DataplaneResource))
		if err != nil {
			return nil, err
		}
		dp, ok := res.(*core_mesh.DataplaneResource)
		if !ok {
			return nil, errors.Errorf("invalid resource")
		}
		if err := dp.Validate(); err != nil {
			return nil, err
		}
		// this part of validation works only for Universal scenarios with TransparentProxying
		if dp.Spec.Networking.TransparentProxying != nil && len(dp.Spec.Networking.Outbound) != 0 {
			var err validators.ValidationError
			err.AddViolation("outbound", "should be empty since dataplane is in Transparent Proxying mode")
			return nil, err.OrNil()
		}
		if err := b.validateMeshExist(ctx, dp.Meta.GetMesh()); err != nil {
			return nil, err
		}
		return dp, nil
	} else {
		dataplane := &core_mesh.DataplaneResource{}
		if err := b.resManager.Get(ctx, dataplane, core_store.GetBy(proxyId.ToResourceKey())); err != nil {
			return nil, err
		}
		return dataplane, nil
	}
}

func (b *bootstrapGenerator) validateMeshExist(ctx context.Context, mesh string) error {
	if err := b.resManager.Get(ctx, &core_mesh.MeshResource{}, core_store.GetByKey(mesh, mesh)); err != nil {
		if core_store.IsResourceNotFound(err) {
			verr := validators.ValidationError{}
			verr.AddViolation("mesh", fmt.Sprintf("mesh %q does not exist", mesh))
			return verr.OrNil()
		}
		return err
	}
	return nil
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

	var certBytes = ""
	if b.xdsCertFile != "" {
		cert, err := ioutil.ReadFile(b.xdsCertFile)
		if err != nil {
			return nil, err
		}
		certBytes = base64.StdEncoding.EncodeToString(cert)
	}
	accessLogPipe := fmt.Sprintf("/tmp/kuma-access-logs-%s-%s.sock", request.Name, request.Mesh)
	params := configParameters{
		Id:                 proxyId.String(),
		Service:            service,
		Zone:               b.zone,
		AdminAddress:       b.config.AdminAddress,
		AdminPort:          adminPort,
		AdminAccessLogPath: b.config.AdminAccessLogPath,
		XdsHost:            b.config.XdsHost,
		XdsPort:            b.config.XdsPort,
		XdsConnectTimeout:  b.config.XdsConnectTimeout,
		AccessLogPipe:      accessLogPipe,
		DataplaneTokenPath: request.DataplaneTokenPath,
		DataplaneResource:  request.DataplaneResource,
		CertBytes:          certBytes,
	}
	log.WithValues("params", params).Info("Generating bootstrap config")
	return b.configForParameters(params)
}

func (b *bootstrapGenerator) verifyAdminPort(adminPort uint32, dataplane *core_mesh.DataplaneResource) error {
	// The admin port in kuma-dp is always bound to 127.0.0.1
	if dataplane.UsesInboundInterface(core_mesh.IPv4Loopback, adminPort) {
		return errors.Errorf("Resource precondition failed: Port %d requested as both admin and inbound port.", adminPort)
	}

	if dataplane.UsesOutboundInterface(core_mesh.IPv4Loopback, adminPort) {
		return errors.Errorf("Resource precondition failed: Port %d requested as both admin and outbound port.", adminPort)
	}
	return nil
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
