package inspect

import (
	"context"
	"encoding/json"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/josephburnett/jd/v2"

	api_common "github.com/kumahq/kuma/api/openapi/types/common"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/test/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/cache/cla"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/generator"
	xds_hooks "github.com/kumahq/kuma/pkg/xds/hooks"
	v3 "github.com/kumahq/kuma/pkg/xds/server/v3"
	"github.com/kumahq/kuma/pkg/xds/sync"
)

type ProxyConfig map[string]interface{}

type ProxyConfigInspector struct {
	zone              string
	meshContext       xds_context.MeshContext
	snapshotGenerator *v3.TemplateSnapshotGenerator
}

func NewProxyConfigInspector(meshContext xds_context.MeshContext, zone string, hooks ...xds_hooks.ResourceSetHook) (*ProxyConfigInspector, error) {
	return &ProxyConfigInspector{
		zone:        zone,
		meshContext: meshContext,
		snapshotGenerator: &v3.TemplateSnapshotGenerator{
			ResourceSetHooks:      hooks,
			ProxyTemplateResolver: generator.DefaultTemplateResolver,
		},
	}, nil
}

func (p *ProxyConfigInspector) Get(ctx context.Context, name string, shadow bool) (ProxyConfig, error) {
	proxyBuilder := &sync.DataplaneProxyBuilder{
		Zone:          p.zone,
		APIVersion:    envoy.APIV3,
		IncludeShadow: shadow,
	}

	proxy, err := proxyBuilder.Build(ctx, model.ResourceKey{Name: name, Mesh: p.mesh()}, p.meshContext)
	if err != nil {
		return nil, err
	}

	envoyCtx := xds_context.Context{
		Mesh: p.meshContext,
		ControlPlane: &xds_context.ControlPlaneContext{
			CLACache: &cla.Retriever{},
			Secrets:  &xds.TestSecrets{},
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
