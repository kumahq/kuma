package builders

import (
	"context"

	hostnamegenerator_api "github.com/kumahq/kuma/pkg/core/resources/apis/hostnamegenerator/api/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

type HostnameGeneratorBuilder struct {
	res *hostnamegenerator_api.HostnameGeneratorResource
}

func HostnameGenerator() *HostnameGeneratorBuilder {
	return &HostnameGeneratorBuilder{
		res: &hostnamegenerator_api.HostnameGeneratorResource{
			Meta: &test_model.ResourceMeta{
				Mesh: core_model.NoMesh,
				Name: "default",
			},
			Spec: &hostnamegenerator_api.HostnameGenerator{},
		},
	}
}

func (m *HostnameGeneratorBuilder) WithLabels(labels map[string]string) *HostnameGeneratorBuilder {
	m.res.Meta.(*test_model.ResourceMeta).Labels = labels
	return m
}

func (m *HostnameGeneratorBuilder) WithName(name string) *HostnameGeneratorBuilder {
	m.res.Meta.(*test_model.ResourceMeta).Name = name
	return m
}

func (m *HostnameGeneratorBuilder) WithTemplate(template string) *HostnameGeneratorBuilder {
	m.res.Spec.Template = template
	return m
}

func (m *HostnameGeneratorBuilder) WithMeshServiceMatchLabels(labels map[string]string) *HostnameGeneratorBuilder {
	m.res.Spec.Selector = hostnamegenerator_api.Selector{
		MeshService: &hostnamegenerator_api.LabelSelector{
			MatchLabels: &labels,
		},
	}
	return m
}

func (m *HostnameGeneratorBuilder) WithMeshExternalServiceMatchLabels(labels map[string]string) *HostnameGeneratorBuilder {
	m.res.Spec.Selector = hostnamegenerator_api.Selector{
		MeshExternalService: &hostnamegenerator_api.LabelSelector{
			MatchLabels: &labels,
		},
	}
	return m
}

func (m *HostnameGeneratorBuilder) WithMeshMultiZoneServiceMatchLabels(labels map[string]string) *HostnameGeneratorBuilder {
	m.res.Spec.Selector = hostnamegenerator_api.Selector{
		MeshMultiZoneService: &hostnamegenerator_api.LabelSelector{
			MatchLabels: &labels,
		},
	}
	return m
}

func (m *HostnameGeneratorBuilder) Build() *hostnamegenerator_api.HostnameGeneratorResource {
	if err := m.res.Validate(); err != nil {
		panic(err)
	}
	return m.res
}

func (m *HostnameGeneratorBuilder) Create(s store.ResourceStore) error {
	opts := []store.CreateOptionsFunc{
		store.CreateBy(m.Key()),
	}
	if ls := m.res.GetMeta().GetLabels(); len(ls) > 0 {
		opts = append(opts, store.CreateWithLabels(ls))
	}
	return s.Create(context.Background(), m.Build(), opts...)
}

func (m *HostnameGeneratorBuilder) Key() core_model.ResourceKey {
	return core_model.MetaToResourceKey(m.res.GetMeta())
}
