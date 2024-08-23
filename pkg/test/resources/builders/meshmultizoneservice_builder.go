package builders

import (
	"context"

	"k8s.io/apimachinery/pkg/util/intstr"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshmzservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

type MeshMultiZoneServiceBuilder struct {
	res *meshmzservice_api.MeshMultiZoneServiceResource
}

func MeshMultiZoneService() *MeshMultiZoneServiceBuilder {
	return &MeshMultiZoneServiceBuilder{
		res: &meshmzservice_api.MeshMultiZoneServiceResource{
			Meta: &test_model.ResourceMeta{
				Mesh: core_model.DefaultMesh,
				Name: "backend",
			},
			Spec:   &meshmzservice_api.MeshMultiZoneService{},
			Status: &meshmzservice_api.MeshMultiZoneServiceStatus{},
		},
	}
}

func (m *MeshMultiZoneServiceBuilder) WithLabels(labels map[string]string) *MeshMultiZoneServiceBuilder {
	m.res.Meta.(*test_model.ResourceMeta).Labels = labels
	return m
}

func (m *MeshMultiZoneServiceBuilder) WithName(name string) *MeshMultiZoneServiceBuilder {
	m.res.Meta.(*test_model.ResourceMeta).Name = name
	return m
}

func (m *MeshMultiZoneServiceBuilder) WithMesh(mesh string) *MeshMultiZoneServiceBuilder {
	m.res.Meta.(*test_model.ResourceMeta).Mesh = mesh
	return m
}

func (m *MeshMultiZoneServiceBuilder) WithServiceLabelSelector(labels map[string]string) *MeshMultiZoneServiceBuilder {
	m.res.Spec.Selector.MeshService.MatchLabels = labels
	return m
}

func (m *MeshMultiZoneServiceBuilder) AddMatchedMeshServiceName(msID core_model.ResourceIdentifier) *MeshMultiZoneServiceBuilder {
	m.res.Status.MeshServices = append(m.res.Status.MeshServices, meshmzservice_api.MatchedMeshService{
		Name:      msID.Name,
		Namespace: msID.Namespace,
		Zone:      msID.Zone,
		Mesh:      msID.Mesh,
	})
	return m
}

func (m *MeshMultiZoneServiceBuilder) AddPort(port meshservice_api.Port) *MeshMultiZoneServiceBuilder {
	m.res.Spec.Ports = append(m.res.Spec.Ports, port)
	return m
}

func (m *MeshMultiZoneServiceBuilder) AddIntPort(port, target uint32, protocol core_mesh.Protocol) *MeshMultiZoneServiceBuilder {
	m.res.Spec.Ports = append(m.res.Spec.Ports, meshservice_api.Port{
		Port: port,
		TargetPort: intstr.IntOrString{
			Type:   intstr.Int,
			IntVal: int32(target),
		},
		AppProtocol: protocol,
	})
	return m
}

func (m *MeshMultiZoneServiceBuilder) Build() *meshmzservice_api.MeshMultiZoneServiceResource {
	if err := m.res.Validate(); err != nil {
		panic(err)
	}
	return m.res
}

func (m *MeshMultiZoneServiceBuilder) Create(s store.ResourceStore) error {
	opts := []store.CreateOptionsFunc{
		store.CreateBy(m.Key()),
	}
	if ls := m.res.GetMeta().GetLabels(); len(ls) > 0 {
		opts = append(opts, store.CreateWithLabels(ls))
	}
	return s.Create(context.Background(), m.Build(), opts...)
}

func (m *MeshMultiZoneServiceBuilder) Key() core_model.ResourceKey {
	return core_model.MetaToResourceKey(m.res.GetMeta())
}
