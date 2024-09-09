package builders

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	mtp_proto "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

type MeshTrafficPermissionBuilder struct {
	res *mtp_proto.MeshTrafficPermissionResource
}

func MeshTrafficPermission() *MeshTrafficPermissionBuilder {
	return &MeshTrafficPermissionBuilder{
		res: &mtp_proto.MeshTrafficPermissionResource{
			Meta: &test_model.ResourceMeta{
				Mesh: core_model.DefaultMesh,
				Name: "mtp-1",
			},
			Spec: &mtp_proto.MeshTrafficPermission{},
		},
	}
}

func (m *MeshTrafficPermissionBuilder) WithName(name string) *MeshTrafficPermissionBuilder {
	m.res.Meta.(*test_model.ResourceMeta).Name = name
	return m
}

func (m *MeshTrafficPermissionBuilder) WithMesh(mesh string) *MeshTrafficPermissionBuilder {
	m.res.Meta.(*test_model.ResourceMeta).Mesh = mesh
	return m
}

func (m *MeshTrafficPermissionBuilder) WithTargetRef(targetRef common_api.TargetRef) *MeshTrafficPermissionBuilder {
	m.res.Spec.TargetRef = &targetRef
	return m
}

func (m *MeshTrafficPermissionBuilder) AddFrom(targetRef common_api.TargetRef, action mtp_proto.Action) *MeshTrafficPermissionBuilder {
	m.res.Spec.From = append(m.res.Spec.From, mtp_proto.From{
		TargetRef: targetRef,
		Default: mtp_proto.Conf{
			Action: action,
		},
	})
	return m
}

func (m *MeshTrafficPermissionBuilder) Build() *mtp_proto.MeshTrafficPermissionResource {
	if err := m.res.Validate(); err != nil {
		panic(err)
	}
	return m.res
}
