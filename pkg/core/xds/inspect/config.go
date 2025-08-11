package inspect

import (
	"context"
	"encoding/json"
	"time"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/josephburnett/jd/v2"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	api_common "github.com/kumahq/kuma/api/openapi/types/common"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/cache/cla"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/generator"
	xds_hooks "github.com/kumahq/kuma/pkg/xds/hooks"
	"github.com/kumahq/kuma/pkg/xds/secrets"
	v3 "github.com/kumahq/kuma/pkg/xds/server/v3"
	"github.com/kumahq/kuma/pkg/xds/sync"
)

type ProxyConfig map[string]interface{}

type ProxyConfigInspector struct {
	zone                   string
	meshContext            xds_context.MeshContext
	snapshotGenerator      *v3.TemplateSnapshotGenerator
	knownInternalAddresses []string
}

func NewProxyConfigInspector(meshContext xds_context.MeshContext, zone string, knownInternalAddresses []string, hooks ...xds_hooks.ResourceSetHook) (*ProxyConfigInspector, error) {
	return &ProxyConfigInspector{
		zone:        zone,
		meshContext: meshContext,
		snapshotGenerator: &v3.TemplateSnapshotGenerator{
			ResourceSetHooks:      hooks,
			ProxyTemplateResolver: generator.DefaultTemplateResolver,
		},
		knownInternalAddresses: knownInternalAddresses,
	}, nil
}

func (p *ProxyConfigInspector) Get(ctx context.Context, name string, shadow bool) (ProxyConfig, error) {
	proxyBuilder := &sync.DataplaneProxyBuilder{
		Zone:              p.zone,
		APIVersion:        envoy.APIV3,
		IncludeShadow:     shadow,
		InternalAddresses: core_xds.InternalAddressesFromCIDRs(p.knownInternalAddresses),
	}

	proxy, err := proxyBuilder.Build(ctx, model.ResourceKey{Name: name, Mesh: p.mesh()}, p.meshContext)
	if err != nil {
		return nil, err
	}

	envoyCtx := xds_context.Context{
		Mesh: p.meshContext,
		ControlPlane: &xds_context.ControlPlaneContext{
			CLACache: &cla.Retriever{},
			Secrets:  &dummySecrets{},
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

func Diff(before, after ProxyConfig) ([]api_common.JsonPatchItem, error) {
	snapshotToNode := func(s map[string]interface{}) (jd.JsonNode, error) {
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
	for rt := 0; rt < int(types.UnknownType); rt++ {
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

// dummySecrets returns hardcoded plausible values
type dummySecrets struct{}

func (ds *dummySecrets) GetForDataPlane(_ context.Context, _ *core_mesh.DataplaneResource, mesh *core_mesh.MeshResource, meshes []*core_mesh.MeshResource) (*core_xds.IdentitySecret, map[string]*core_xds.CaSecret, error) {
	return ds.identity(), ds.cas(append(meshes, mesh)...), nil
}

func (ds *dummySecrets) GetForZoneEgress(_ context.Context, _ *core_mesh.ZoneEgressResource, mesh *core_mesh.MeshResource) (*core_xds.IdentitySecret, *core_xds.CaSecret, error) {
	return ds.identity(), ds.cas(mesh)[mesh.GetMeta().GetName()], nil
}

func (ds *dummySecrets) GetAllInOne(ctx context.Context, _ *core_mesh.MeshResource, _ *core_mesh.DataplaneResource, _ []*core_mesh.MeshResource) (*core_xds.IdentitySecret, *core_xds.CaSecret, error) {
	return ds.identity(), &core_xds.CaSecret{PemCerts: [][]byte{[]byte("COMBINED")}}, nil
}

func (ds *dummySecrets) identity() *core_xds.IdentitySecret {
	return &core_xds.IdentitySecret{PemCerts: [][]byte{[]byte("CERT")}, PemKey: []byte("KEY")}
}

func (ds *dummySecrets) cas(meshes ...*core_mesh.MeshResource) map[string]*core_xds.CaSecret {
	cas := map[string]*core_xds.CaSecret{}
	for _, mesh := range meshes {
		cas[mesh.GetMeta().GetName()] = &core_xds.CaSecret{PemCerts: [][]byte{[]byte("CA")}}
	}
	return cas
}

func (ds *dummySecrets) Info(proxyType mesh_proto.ProxyType, dpKey model.ResourceKey) *secrets.Info {
	return &secrets.Info{
		Expiration: time.Unix(2, 2),
		Generation: time.Unix(1, 1),
		Tags: map[string]map[string]bool{
			"kuma.io/service": {
				dpKey.Name: true,
			},
		},
		OwnMesh: secrets.MeshInfo{
			MTLS: &mesh_proto.Mesh_Mtls{
				EnabledBackend: "ca-1",
				Backends:       nil,
			},
		},
		IssuedBackend:     "ca-1",
		SupportedBackends: []string{"ca-1"},
	}
}

func (ds *dummySecrets) Cleanup(mesh_proto.ProxyType, model.ResourceKey) {}
