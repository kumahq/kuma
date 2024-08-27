package builders

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

type DataplaneInsightBuilder struct {
	res *core_mesh.DataplaneInsightResource
}

func DataplaneInsight() *DataplaneInsightBuilder {
	return &DataplaneInsightBuilder{
		res: &core_mesh.DataplaneInsightResource{
			Meta: &test_model.ResourceMeta{
				Mesh: core_model.DefaultMesh,
				Name: "dp-1",
			},
			Spec: &mesh_proto.DataplaneInsight{},
		},
	}
}

func (d *DataplaneInsightBuilder) Build() *core_mesh.DataplaneInsightResource {
	return d.res
}

func (d *DataplaneInsightBuilder) Create(s store.ResourceStore) error {
	return s.Create(context.Background(), d.Build(), store.CreateBy(d.Key()))
}

func (d *DataplaneInsightBuilder) Key() core_model.ResourceKey {
	return core_model.MetaToResourceKey(d.res.GetMeta())
}

func (d *DataplaneInsightBuilder) With(fn func(resource *core_mesh.DataplaneInsightResource)) *DataplaneInsightBuilder {
	fn(d.res)
	return d
}

func (d *DataplaneInsightBuilder) WithName(name string) *DataplaneInsightBuilder {
	d.res.Meta.(*test_model.ResourceMeta).Name = name
	return d
}

func (d *DataplaneInsightBuilder) WithMesh(mesh string) *DataplaneInsightBuilder {
	d.res.Meta.(*test_model.ResourceMeta).Mesh = mesh
	return d
}

func (d *DataplaneInsightBuilder) WithMTLS(mtls *mesh_proto.DataplaneInsight_MTLS) *DataplaneInsightBuilder {
	d.res.Spec.MTLS = mtls
	return d
}

func (d *DataplaneInsightBuilder) AddSubscription(sub *mesh_proto.DiscoverySubscription) *DataplaneInsightBuilder {
	d.res.Spec.Subscriptions = append(d.res.Spec.Subscriptions, sub)
	return d
}

func (d *DataplaneInsightBuilder) WithMTLSIssuedBackend(issuedBackend string) *DataplaneInsightBuilder {
	if d.res.Spec.MTLS == nil {
		d.res.Spec.MTLS = &mesh_proto.DataplaneInsight_MTLS{}
	}
	d.res.Spec.MTLS.IssuedBackend = issuedBackend
	return d
}
