package builders

import (
	"context"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

type MeshExternalServiceBuilder struct {
	res *v1alpha1.MeshExternalServiceResource
}

func MeshExternalService() *MeshExternalServiceBuilder {
	return &MeshExternalServiceBuilder{
		res: &v1alpha1.MeshExternalServiceResource{
			Meta: &test_model.ResourceMeta{
				Mesh: core_model.DefaultMesh,
				Name: "example",
			},
			Spec: &v1alpha1.MeshExternalService{
				Match: v1alpha1.Match{
					Type:     pointer.To(v1alpha1.HostnameGeneratorType),
					Port:     9000,
					Protocol: core_mesh.ProtocolHTTP,
				},
				Endpoints: []v1alpha1.Endpoint{
					{
						Address: "192.168.0.1",
						Port:    pointer.To[v1alpha1.Port](27017),
					},
				},
			},
			Status: &v1alpha1.MeshExternalServiceStatus{},
		},
	}
}

func (m *MeshExternalServiceBuilder) WithLabels(labels map[string]string) *MeshExternalServiceBuilder {
	m.res.Meta.(*test_model.ResourceMeta).Labels = labels
	return m
}

func (m *MeshExternalServiceBuilder) WithName(name string) *MeshExternalServiceBuilder {
	m.res.Meta.(*test_model.ResourceMeta).Name = name
	return m
}

func (m *MeshExternalServiceBuilder) WithMesh(mesh string) *MeshExternalServiceBuilder {
	m.res.Meta.(*test_model.ResourceMeta).Mesh = mesh
	return m
}

func (m *MeshExternalServiceBuilder) WithKumaVIP(vip string) *MeshExternalServiceBuilder {
	m.res.Status.VIP = v1alpha1.VIP{IP: vip}
	return m
}

func (m *MeshExternalServiceBuilder) WithoutVIP() *MeshExternalServiceBuilder {
	m.res.Status.VIP.IP = ""
	return m
}

func (m *MeshExternalServiceBuilder) Build() *v1alpha1.MeshExternalServiceResource {
	if err := m.res.Validate(); err != nil {
		panic(err)
	}
	return m.res
}

func (m *MeshExternalServiceBuilder) Create(s store.ResourceStore) error {
	opts := []store.CreateOptionsFunc{
		store.CreateBy(m.Key()),
	}
	if ls := m.res.GetMeta().GetLabels(); len(ls) > 0 {
		opts = append(opts, store.CreateWithLabels(ls))
	}
	return s.Create(context.Background(), m.Build(), opts...)
}

func (m *MeshExternalServiceBuilder) Key() core_model.ResourceKey {
	return core_model.MetaToResourceKey(m.res.GetMeta())
}
