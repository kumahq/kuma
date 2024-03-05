package builders

import (
	"context"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

type ZoneInsightBuilder struct {
	res *system.ZoneInsightResource
}

func ZoneInsight() *ZoneInsightBuilder {
	return &ZoneInsightBuilder{
		res: &system.ZoneInsightResource{
			Meta: &test_model.ResourceMeta{
				Mesh: core_model.NoMesh,
				Name: "zoneInsight-1",
			},
			Spec: &system_proto.ZoneInsight{
				Subscriptions: []*system_proto.KDSSubscription{},
			},
		},
	}
}

func (zi *ZoneInsightBuilder) Build() *system.ZoneInsightResource {
	return zi.res
}

func (zi *ZoneInsightBuilder) Create(s store.ResourceStore) error {
	return s.Create(context.Background(), zi.Build(), store.CreateBy(zi.Key()))
}

func (zi *ZoneInsightBuilder) Key() core_model.ResourceKey {
	return core_model.MetaToResourceKey(zi.res.GetMeta())
}

func (zi *ZoneInsightBuilder) WithName(name string) *ZoneInsightBuilder {
	zi.res.Meta.(*test_model.ResourceMeta).Name = name
	return zi
}

func (zi *ZoneInsightBuilder) AddSubscription(subscription *system_proto.KDSSubscription) *ZoneInsightBuilder {
	zi.res.Spec.Subscriptions = append(zi.res.Spec.Subscriptions, subscription)
	return zi
}
