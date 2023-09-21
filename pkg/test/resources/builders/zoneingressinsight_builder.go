package builders

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

type ZoneIngressInsightBuilder struct {
	res *mesh.ZoneIngressInsightResource
}

func ZoneIngressInsight() *ZoneIngressInsightBuilder {
	return &ZoneIngressInsightBuilder{
		res: &mesh.ZoneIngressInsightResource{
			Meta: &test_model.ResourceMeta{
				Mesh: core_model.NoMesh,
				Name: "zoneIngressInsight-1",
			},
			Spec: &mesh_proto.ZoneIngressInsight{
				Subscriptions: []*mesh_proto.DiscoverySubscription{},
			},
		},
	}
}

func (zii *ZoneIngressInsightBuilder) Build() *mesh.ZoneIngressInsightResource {
	return zii.res
}

func (zii *ZoneIngressInsightBuilder) Create(s store.ResourceStore) error {
	return s.Create(context.Background(), zii.Build(), store.CreateBy(zii.Key()))
}

func (zii *ZoneIngressInsightBuilder) Key() core_model.ResourceKey {
	return core_model.MetaToResourceKey(zii.res.GetMeta())
}

func (zii *ZoneIngressInsightBuilder) WithName(name string) *ZoneIngressInsightBuilder {
	zii.res.Meta.(*test_model.ResourceMeta).Name = name
	return zii
}

func (zii *ZoneIngressInsightBuilder) WithMesh(mesh string) *ZoneIngressInsightBuilder {
	zii.res.Meta.(*test_model.ResourceMeta).Mesh = mesh
	return zii
}

func (zii *ZoneIngressInsightBuilder) AddSubscription(subscription *mesh_proto.DiscoverySubscription) *ZoneIngressInsightBuilder {
	zii.res.Spec.Subscriptions = append(zii.res.Spec.Subscriptions, subscription)
	return zii
}
