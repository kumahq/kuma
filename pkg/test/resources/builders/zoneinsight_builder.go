package builders

import (
	"context"

	zoneinsight_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/zoneinsight/api/v1alpha1"

	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
)

type ZoneInsightBuilder struct {
	res *zoneinsight_api.ZoneInsightResource
}

func ZoneInsight() *ZoneInsightBuilder {
	return &ZoneInsightBuilder{
		res: &zoneinsight_api.ZoneInsightResource{
			Meta: &test_model.ResourceMeta{
				Mesh: core_model.NoMesh,
				Name: "zoneInsight-1",
			},
			Spec: &zoneinsight_api.ZoneInsight{
				Subscriptions: []*zoneinsight_api.KDSSubscription{},
			},
		},
	}
}

func (zi *ZoneInsightBuilder) Build() *zoneinsight_api.ZoneInsightResource {
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

func (zi *ZoneInsightBuilder) AddSubscription(subscription *zoneinsight_api.KDSSubscription) *ZoneInsightBuilder {
	zi.res.Spec.Subscriptions = append(zi.res.Spec.Subscriptions, subscription)
	return zi
}
