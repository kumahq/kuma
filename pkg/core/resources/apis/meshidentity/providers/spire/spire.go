package spire

import (
	"context"
	"time"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/go-logr/logr"
	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"

	config_core "github.com/kumahq/kuma/v2/pkg/config/core"
	"github.com/kumahq/kuma/v2/pkg/core"
	"github.com/kumahq/kuma/v2/pkg/core/kri"
	meshidentity_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshidentity/providers"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshidentity/providers/spire/metadata"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	core_system_names "github.com/kumahq/kuma/v2/pkg/core/system_names"
	"github.com/kumahq/kuma/v2/pkg/core/xds"
	"github.com/kumahq/kuma/v2/pkg/core/xds/types"
	bldrs_cluster "github.com/kumahq/kuma/v2/pkg/envoy/builders/cluster"
	bldrs_common "github.com/kumahq/kuma/v2/pkg/envoy/builders/common"
	bldrs_core "github.com/kumahq/kuma/v2/pkg/envoy/builders/core"
	bldrs_tls "github.com/kumahq/kuma/v2/pkg/envoy/builders/tls"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
)

const (
	// Secret name which includes all CAs required after federation
	FederatedCASecretName        = "ALL"
	defaultSpireAgentConnTimeout = 1 * time.Second
)

var SpireAgentClusterName = core_system_names.AsSystemName("identity_sds-spire-agent")

var _ providers.IdentityProvider = &spireIdentityProvider{}

type spireIdentityProvider struct {
	logger      logr.Logger
	socketPath  string
	zone        string
	environment config_core.EnvironmentType
}

func NewSpireIdentityProvider(socketPath, zone string, environment config_core.EnvironmentType) providers.IdentityProvider {
	logger := core.Log.WithName("identity-provider").WithName("spire")
	return &spireIdentityProvider{
		logger:      logger,
		socketPath:  socketPath,
		zone:        zone,
		environment: environment,
	}
}

func (s *spireIdentityProvider) Validate(ctx context.Context, identity *meshidentity_api.MeshIdentityResource) error {
	return nil
}

func (s *spireIdentityProvider) Initialize(ctx context.Context, identity *meshidentity_api.MeshIdentityResource) error {
	return nil
}

// All certificates configuration is handled by the Spire
func (s *spireIdentityProvider) GetRootCA(ctx context.Context, identity *meshidentity_api.MeshIdentityResource) ([]byte, error) {
	return nil, nil
}

func (s *spireIdentityProvider) ShouldCreateMeshTrust(_ *meshidentity_api.MeshIdentityResource) bool {
	return false
}

func (s *spireIdentityProvider) CreateIdentity(ctx context.Context, identity *meshidentity_api.MeshIdentityResource, proxy *xds.Proxy) (*xds.WorkloadIdentity, error) {
	if s.environment == config_core.KubernetesEnvironment && !proxy.Metadata.HasFeature(types.FeatureSpire) {
		s.logger.Info("dataplane doesn't have spire socket mounted, please redeploy your Pod", "dpp", model.MetaToResourceKey(proxy.Dataplane.GetMeta()), "identity", model.MetaToResourceKey(identity.GetMeta()))
		return nil, nil
	}
	trustDomain, err := identity.Spec.GetTrustDomain(identity.GetMeta(), s.zone)
	if err != nil {
		return nil, err
	}

	spiffeID, err := identity.Spec.GetSpiffeID(trustDomain, proxy.Dataplane.GetMeta(), s.environment)
	if err != nil {
		return nil, err
	}
	connectTimeout := k8s.Duration{Duration: defaultSpireAgentConnTimeout}
	if identity.Spec.Provider.Spire != nil && identity.Spec.Provider.Spire.Agent != nil {
		connectTimeout = pointer.DerefOr(identity.Spec.Provider.Spire.Agent.Timeout, k8s.Duration{Duration: defaultSpireAgentConnTimeout})
	}
	socketPath := s.socketPath
	if proxy.Metadata != nil && proxy.Metadata.SpireSocketPath != "" {
		socketPath = proxy.Metadata.SpireSocketPath
	}
	resources, err := additionalResources(socketPath, connectTimeout.Duration)
	if err != nil {
		return nil, err
	}

	return &xds.WorkloadIdentity{
		KRI:                                kri.From(identity),
		ManagementMode:                     xds.ExternalManagementMode,
		IdentitySourceConfigurer:           sourceConfigurer(spiffeID),
		ExternalValidationSourceConfigurer: sourceConfigurer(FederatedCASecretName),
		AdditionalResources:                resources,
	}, nil
}

// we need to create a cluster for spire agent
func additionalResources(socketPath string, timeout time.Duration) (*xds.ResourceSet, error) {
	resources := xds.NewResourceSet()
	resource, err := bldrs_cluster.NewCluster().
		Configure(bldrs_cluster.Name(SpireAgentClusterName)).
		Configure(bldrs_cluster.ConnectTimeout(timeout)).
		Configure(bldrs_cluster.Http2()).
		Configure(bldrs_cluster.Endpoints(SpireAgentClusterName, []*envoy_endpoint.LocalityLbEndpoints{
			{
				LbEndpoints: []*envoy_endpoint.LbEndpoint{
					{
						HostIdentifier: &envoy_endpoint.LbEndpoint_Endpoint{
							Endpoint: &envoy_endpoint.Endpoint{
								Address: &envoy_core.Address{
									Address: &envoy_core.Address_Pipe{
										Pipe: &envoy_core.Pipe{
											Path: socketPath,
										},
									},
								},
							},
						},
					},
				},
			},
		})).Build()
	if err != nil {
		return nil, err
	}
	resources = resources.Add(&xds.Resource{
		Name:     SpireAgentClusterName,
		Origin:   metadata.OriginMeshTrust,
		Resource: resource,
	})
	return resources, nil
}

func sourceConfigurer(secretName string) func() bldrs_common.Configurer[envoy_tls.SdsSecretConfig] {
	return func() bldrs_common.Configurer[envoy_tls.SdsSecretConfig] {
		return bldrs_tls.SdsSecretConfigSource(
			secretName,
			bldrs_core.NewConfigSource().Configure(bldrs_core.ApiConfigSource(SpireAgentClusterName)),
		)
	}
}
