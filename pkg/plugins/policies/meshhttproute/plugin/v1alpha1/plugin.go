package v1alpha1

import (
	"github.com/pkg/errors"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/matchers"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/subsetutils"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds/meshroute"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
)

var _ core_plugins.PolicyPlugin = &plugin{}

type ToRouteRule struct {
	Subset    subsetutils.Subset
	Rules     []api.Rule
	Hostnames []string

	Origins          []core_model.ResourceMeta
	BackendRefOrigin map[common_api.MatchesHash]core_model.ResourceMeta
}

type plugin struct{}

func (p plugin) Order() int { return api.MeshHTTPRouteResourceTypeDescriptor.Order }

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources, opts ...core_plugins.MatchedPoliciesOption) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshHTTPRouteType, dataplane, resources, opts...)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, xdsCtx xds_context.Context, proxy *core_xds.Proxy) error {
	if proxy.Dataplane == nil {
		return nil
	}

	// These policies have already been merged using the custom `GetDefault`
	// method and therefore are of the
	// `ToRouteRule` type, where rules have been appended together.
	policies := proxy.Policies.Dynamic[api.MeshHTTPRouteType]

	if err := ApplyToOutbounds(proxy, rs, xdsCtx, policies.ToRules); err != nil {
		return err
	}

	return nil
}

func ApplyToOutbounds(proxy *core_xds.Proxy, rs *core_xds.ResourceSet, xdsCtx xds_context.Context, rules rules.ToRules) error {
	tlsReady := xdsCtx.Mesh.GetTLSReadiness()
	servicesAcc := envoy_common.NewServicesAccumulator(tlsReady)

	listeners, err := generateListeners(proxy, rules, servicesAcc, xdsCtx.Mesh)
	if err != nil {
		return errors.Wrap(err, "couldn't generate listener resources")
	}
	rs.AddSet(listeners)

	services := servicesAcc.Services()

	clusters, err := meshroute.GenerateClusters(proxy, xdsCtx.Mesh, services, xdsCtx.ControlPlane.SystemNamespace)
	if err != nil {
		return errors.Wrap(err, "couldn't generate cluster resources")
	}
	rs.AddSet(clusters)

	endpoints, err := meshroute.GenerateEndpoints(proxy, xdsCtx, services)
	if err != nil {
		return errors.Wrap(err, "couldn't generate endpoint resources")
	}
	rs.AddSet(endpoints)

	return nil
}
