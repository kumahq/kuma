package builders

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

type ServiceInsightBuilder struct {
	res *mesh.ServiceInsightResource
}

func ServiceInsight() *ServiceInsightBuilder {
	return &ServiceInsightBuilder{
		res: &mesh.ServiceInsightResource{
			Meta: &test_model.ResourceMeta{
				Mesh: core_model.DefaultMesh,
				Name: "serviceInsight-1",
			},
			Spec: &mesh_proto.ServiceInsight{
				Services: map[string]*mesh_proto.ServiceInsight_Service{},
			},
		},
	}
}

func (si *ServiceInsightBuilder) Build() *mesh.ServiceInsightResource {
	return si.res
}

func (si *ServiceInsightBuilder) Create(s store.ResourceStore) error {
	return s.Create(context.Background(), si.Build(), store.CreateBy(si.Key()))
}

func (si *ServiceInsightBuilder) Key() core_model.ResourceKey {
	return core_model.MetaToResourceKey(si.res.GetMeta())
}

func (si *ServiceInsightBuilder) WithName(name string) *ServiceInsightBuilder {
	si.res.Meta.(*test_model.ResourceMeta).Name = name
	return si
}

func (si *ServiceInsightBuilder) WithMesh(mesh string) *ServiceInsightBuilder {
	si.res.Meta.(*test_model.ResourceMeta).Mesh = mesh
	return si
}

func (si *ServiceInsightBuilder) AddService(name string, service *mesh_proto.ServiceInsight_Service) *ServiceInsightBuilder {
	si.res.Spec.Services[name] = service
	return si
}
