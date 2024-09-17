package builders

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/util/proto"
)

type MeshBuilder struct {
	res *core_mesh.MeshResource
}

func Mesh() *MeshBuilder {
	return &MeshBuilder{
		res: &core_mesh.MeshResource{
			Meta: &test_model.ResourceMeta{
				Mesh: core_model.NoMesh,
				Name: core_model.DefaultMesh,
			},
			Spec: &mesh_proto.Mesh{},
		},
	}
}

func (m *MeshBuilder) Build() *core_mesh.MeshResource {
	if err := m.res.Validate(); err != nil {
		panic(err)
	}
	return m.res
}

func (m *MeshBuilder) Create(s store.ResourceStore) error {
	return s.Create(context.Background(), m.Build(), store.CreateBy(m.Key()))
}

func (m *MeshBuilder) Key() core_model.ResourceKey {
	return core_model.MetaToResourceKey(m.res.GetMeta())
}

func (m *MeshBuilder) WithName(name string) *MeshBuilder {
	m.res.Meta.(*test_model.ResourceMeta).Name = name
	return m
}

func (m *MeshBuilder) WithEnabledMTLSBackend(name string) *MeshBuilder {
	if m.res.Spec.Mtls == nil {
		m.res.Spec.Mtls = &mesh_proto.Mesh_Mtls{}
	}
	m.res.Spec.Mtls.EnabledBackend = name
	return m
}

func (m *MeshBuilder) WithBuiltinMTLSBackend(name string) *MeshBuilder {
	if m.res.Spec.Mtls == nil {
		m.res.Spec.Mtls = &mesh_proto.Mesh_Mtls{}
	}
	return m.AddBuiltinMTLSBackend(name)
}

func (m *MeshBuilder) WithoutBackendValidation() *MeshBuilder {
	if m.res.Spec.Mtls == nil {
		m.res.Spec.Mtls = &mesh_proto.Mesh_Mtls{}
	}
	m.res.Spec.Mtls.SkipValidation = true
	return m
}

func (m *MeshBuilder) WithoutMTLSBackends() *MeshBuilder {
	m.res.Spec.Mtls.Backends = nil
	return m
}

func (m *MeshBuilder) WithPermissiveMTLSBackends() *MeshBuilder {
	for _, backend := range m.res.Spec.Mtls.Backends {
		backend.Mode = mesh_proto.CertificateAuthorityBackend_PERMISSIVE
	}
	return m
}

func (m *MeshBuilder) AddBuiltinMTLSBackend(name string) *MeshBuilder {
	if m.res.Spec.Mtls == nil {
		m.res.Spec.Mtls = &mesh_proto.Mesh_Mtls{}
	}
	m.res.Spec.Mtls.Backends = append(m.res.Spec.Mtls.Backends, &mesh_proto.CertificateAuthorityBackend{
		Name: name,
		Type: "builtin",
	})
	return m
}

func (m *MeshBuilder) WithEgressRoutingEnabled() *MeshBuilder {
	if m.res.Spec.Routing == nil {
		m.res.Spec.Routing = &mesh_proto.Routing{}
	}
	m.res.Spec.Routing.ZoneEgress = true
	return m
}

func (m *MeshBuilder) WithMeshExternalServiceTrafficForbidden() *MeshBuilder {
	if m.res.Spec.Routing == nil {
		m.res.Spec.Routing = &mesh_proto.Routing{}
	}
	m.res.Spec.Routing.DefaultForbidMeshExternalServiceAccess = true
	return m
}

func (m *MeshBuilder) WithoutPassthrough() *MeshBuilder {
	if m.res.Spec.Networking == nil {
		m.res.Spec.Networking = &mesh_proto.Networking{}
	}
	if m.res.Spec.Networking.Outbound == nil {
		m.res.Spec.Networking.Outbound = &mesh_proto.Networking_Outbound{}
	}
	m.res.Spec.Networking.Outbound.Passthrough = proto.Bool(false)
	return m
}

func (m *MeshBuilder) WithMeshServicesEnabled(enabled mesh_proto.Mesh_MeshServices_Enabled) *MeshBuilder {
	m.res.Spec.MeshServices = &mesh_proto.Mesh_MeshServices{
		Enabled: enabled,
	}
	return m
}

func (m *MeshBuilder) With(fn func(resource *core_mesh.MeshResource)) *MeshBuilder {
	fn(m.res)
	return m
}

func (m *MeshBuilder) KubeYaml() string {
	mesh := m.Build()
	kubeMesh := mesh_k8s.Mesh{
		TypeMeta: v1.TypeMeta{
			Kind:       string(core_mesh.MeshType),
			APIVersion: mesh_k8s.GroupVersion.String(),
		},
		ObjectMeta: v1.ObjectMeta{
			Name: mesh.Meta.GetName(),
		},
	}
	kubeMesh.SetSpec(mesh.Spec)
	res, err := yaml.Marshal(kubeMesh)
	if err != nil {
		panic(err)
	}
	return string(res)
}
