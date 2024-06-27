package builders

import (
	"context"

	"k8s.io/apimachinery/pkg/util/intstr"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

type MeshServiceBuilder struct {
	res *v1alpha1.MeshServiceResource
}

func MeshService() *MeshServiceBuilder {
	return &MeshServiceBuilder{
		res: &v1alpha1.MeshServiceResource{
			Meta: &test_model.ResourceMeta{
				Mesh: core_model.DefaultMesh,
				Name: "backend",
			},
			Spec:   &v1alpha1.MeshService{},
			Status: &v1alpha1.MeshServiceStatus{},
		},
	}
}

func (m *MeshServiceBuilder) WithLabels(labels map[string]string) *MeshServiceBuilder {
	m.res.Meta.(*test_model.ResourceMeta).Labels = labels
	return m
}

func (m *MeshServiceBuilder) WithName(name string) *MeshServiceBuilder {
	m.res.Meta.(*test_model.ResourceMeta).Name = name
	return m
}

func (m *MeshServiceBuilder) WithMesh(mesh string) *MeshServiceBuilder {
	m.res.Meta.(*test_model.ResourceMeta).Mesh = mesh
	return m
}

func (m *MeshServiceBuilder) WithDataplaneRefNameSelector(name string) *MeshServiceBuilder {
	m.res.Spec.Selector = v1alpha1.Selector{
		DataplaneRef: &v1alpha1.DataplaneRef{
			Name: name,
		},
	}
	return m
}

func (m *MeshServiceBuilder) WithDataplaneTagsSelector(selector map[string]string) *MeshServiceBuilder {
	m.res.Spec.Selector = v1alpha1.Selector{
		DataplaneTags: selector,
	}
	return m
}

func (m *MeshServiceBuilder) WithDataplaneTagsSelectorKV(selectorKV ...string) *MeshServiceBuilder {
	return m.WithDataplaneTagsSelector(TagsKVToMap(selectorKV))
}

func (m *MeshServiceBuilder) AddIntPort(port, target uint32, protocol core_mesh.Protocol) *MeshServiceBuilder {
	m.res.Spec.Ports = append(m.res.Spec.Ports, v1alpha1.Port{
		Port: port,
		TargetPort: intstr.IntOrString{
			Type:   intstr.Int,
			IntVal: int32(target),
		},
		AppProtocol: protocol,
	})
	return m
}

func (m *MeshServiceBuilder) WithKumaVIP(vip string) *MeshServiceBuilder {
	m.res.Status.VIPs = []v1alpha1.VIP{
		{
			IP: vip,
		},
	}
	return m
}

func (m *MeshServiceBuilder) WithoutVIP() *MeshServiceBuilder {
	m.res.Status.VIPs = []v1alpha1.VIP{}
	return m
}

func (m *MeshServiceBuilder) Build() *v1alpha1.MeshServiceResource {
	if err := m.res.Validate(); err != nil {
		panic(err)
	}
	return m.res
}

func (m *MeshServiceBuilder) Create(s store.ResourceStore) error {
	opts := []store.CreateOptionsFunc{
		store.CreateBy(m.Key()),
	}
	if ls := m.res.GetMeta().GetLabels(); len(ls) > 0 {
		opts = append(opts, store.CreateWithLabels(ls))
	}
	return s.Create(context.Background(), m.Build(), opts...)
}

func (m *MeshServiceBuilder) Key() core_model.ResourceKey {
	return core_model.MetaToResourceKey(m.res.GetMeta())
}
