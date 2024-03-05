package builders

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

type ZoneEgressInsightBuilder struct {
	res *mesh.ZoneEgressInsightResource
}

func ZoneEgressInsight() *ZoneEgressInsightBuilder {
	return &ZoneEgressInsightBuilder{
		res: &mesh.ZoneEgressInsightResource{
			Meta: &test_model.ResourceMeta{
				Mesh: core_model.DefaultMesh,
				Name: "zoneIngressInsight-1",
			},
			Spec: &mesh_proto.ZoneEgressInsight{
				Subscriptions: []*mesh_proto.DiscoverySubscription{},
			},
		},
	}
}

func (zei *ZoneEgressInsightBuilder) Build() *mesh.ZoneEgressInsightResource {
	return zei.res
}

func (zei *ZoneEgressInsightBuilder) Create(s store.ResourceStore) error {
	return s.Create(context.Background(), zei.Build(), store.CreateBy(zei.Key()))
}

func (zei *ZoneEgressInsightBuilder) Key() core_model.ResourceKey {
	return core_model.MetaToResourceKey(zei.res.GetMeta())
}

func (zei *ZoneEgressInsightBuilder) WithName(name string) *ZoneEgressInsightBuilder {
	zei.res.Meta.(*test_model.ResourceMeta).Name = name
	return zei
}

func (zei *ZoneEgressInsightBuilder) WithMesh(mesh string) *ZoneEgressInsightBuilder {
	zei.res.Meta.(*test_model.ResourceMeta).Mesh = mesh
	return zei
}

func (zei *ZoneEgressInsightBuilder) AddSubscription(subscription *mesh_proto.DiscoverySubscription) *ZoneEgressInsightBuilder {
	zei.res.Spec.Subscriptions = append(zei.res.Spec.Subscriptions, subscription)
	return zei
}
