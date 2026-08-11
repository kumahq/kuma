package inspect

import (
	"context"
	"encoding/json"

	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/josephburnett/jd/v2"

	api_common "github.com/kumahq/kuma/v3/api/openapi/types/common"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	meshidentity_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	bldrs_common "github.com/kumahq/kuma/v3/pkg/envoy/builders/common"
	bldrs_core "github.com/kumahq/kuma/v3/pkg/envoy/builders/core"
	bldrs_tls "github.com/kumahq/kuma/v3/pkg/envoy/builders/tls"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	"github.com/kumahq/kuma/v3/pkg/xds/cache/cla"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	"github.com/kumahq/kuma/v3/pkg/xds/envoy"
	xds_hooks "github.com/kumahq/kuma/v3/pkg/xds/hooks"
	v3 "github.com/kumahq/kuma/v3/pkg/xds/server/v3"
	"github.com/kumahq/kuma/v3/pkg/xds/sync"
)

type ProxyConfig map[string]any

type ProxyConfigInspector struct {
	zone                   string
	meshContext            xds_context.MeshContext
	snapshotGenerator      *v3.TemplateSnapshotGenerator
	knownInternalAddresses []string
	dataplaneMetadata      *core_xds.DataplaneMetadata
}

func NewProxyConfigInspector(meshContext xds_context.MeshContext, dataplaneMetadata *core_xds.DataplaneMetadata, zone string, knownInternalAddresses []string, hooks ...xds_hooks.ResourceSetHook) (*ProxyConfigInspector, error) {
	return &ProxyConfigInspector{
		zone:        zone,
		meshContext: meshContext,
		snapshotGenerator: &v3.TemplateSnapshotGenerator{
			ResourceSetHooks: hooks,
		},
		knownInternalAddresses: knownInternalAddresses,
		dataplaneMetadata:      dataplaneMetadata,
	}, nil
}

func (p *ProxyConfigInspector) Get(ctx context.Context, name string, shadow bool) (ProxyConfig, error) {
	proxyBuilder := &sync.DataplaneProxyBuilder{
		Zone:              p.zone,
		APIVersion:        envoy.APIV3,
		IncludeShadow:     shadow,
		InternalAddresses: core_xds.InternalAddressesFromCIDRs(p.knownInternalAddresses),
	}

	proxy, err := proxyBuilder.Build(ctx, model.ResourceKey{Name: name, Mesh: p.mesh()}, p.dataplaneMetadata, p.meshContext)
	if err != nil {
		return nil, err
	}

	if identity, _ := meshidentity_api.BestMatched(proxy.Dataplane.GetMeta().GetLabels(), p.meshContext.Resources.MeshIdentities().Items); identity != nil && identity.Status != nil && identity.Status.IsInitialized() {
		proxy.WorkloadIdentity = dummyIdentity(identity)
	}

	envoyCtx := xds_context.Context{
		Mesh: p.meshContext,
		ControlPlane: &xds_context.ControlPlaneContext{
			CLACache: &cla.Retriever{},
			Zone:     p.zone,
		},
	}

	s, err := p.snapshotGenerator.GenerateSnapshot(ctx, envoyCtx, proxy)
	if err != nil {
		return nil, err
	}

	return marshalSnapshot(s)
}

func (p *ProxyConfigInspector) mesh() string {
	return p.meshContext.Resource.GetMeta().GetName()
}

func dummyIdentity(identity *meshidentity_api.MeshIdentityResource) *core_xds.WorkloadIdentity {
	identifier := kri.From(identity)
	secretName := identifier.String()
	return &core_xds.WorkloadIdentity{
		KRI:            identifier,
		ManagementMode: core_xds.KumaManagementMode,
		IdentitySourceConfigurer: func() bldrs_common.Configurer[envoy_tls.SdsSecretConfig] {
			return bldrs_tls.SdsSecretConfigSource(
				secretName,
				bldrs_core.NewConfigSource().Configure(bldrs_core.Sds()),
			)
		},
	}
}

func Diff(before, after ProxyConfig) ([]api_common.JsonPatchItem, error) {
	snapshotToNode := func(s map[string]any) (jd.JsonNode, error) {
		bytes, err := json.Marshal(s)
		if err != nil {
			return nil, err
		}
		return jd.ReadJsonString(string(bytes))
	}

	beforeNode, err := snapshotToNode(before)
	if err != nil {
		return nil, err
	}

	afterNode, err := snapshotToNode(after)
	if err != nil {
		return nil, err
	}

	diff, err := beforeNode.Diff(afterNode).RenderPatch()
	if err != nil {
		return nil, err
	}

	rv := []api_common.JsonPatchItem{}
	if err := json.Unmarshal([]byte(diff), &rv); err != nil {
		return nil, err
	}

	return rv, nil
}

func marshalSnapshot(s *cache.Snapshot) (ProxyConfig, error) {
	resourcesByType := ProxyConfig{}
	for rt := range int(types.UnknownType) {
		items := s.Resources[rt].Items
		if len(items) == 0 {
			continue
		}
		raw, err := marshalMap(items)
		if err != nil {
			return nil, err
		}
		responseType, err := cache.GetResponseTypeURL(types.ResponseType(rt))
		if err != nil {
			return nil, err
		}
		resourcesByType[responseType] = raw
	}
	return resourcesByType, nil
}

func marshalMap(m map[string]types.ResourceWithTTL) (json.RawMessage, error) {
	resourcesByName := make(map[string]json.RawMessage)
	for name, r := range m {
		raw, err := util_proto.ToJSON(r.Resource)
		if err != nil {
			return nil, err
		}
		resourcesByName[name] = raw
	}
	return json.Marshal(resourcesByName)
}
