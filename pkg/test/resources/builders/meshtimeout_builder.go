package builders

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	meshtimeout_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

type MeshTimeoutBuilder struct {
	res *meshtimeout_api.MeshTimeoutResource
}

func MeshTimeout() *MeshTimeoutBuilder {
	return &MeshTimeoutBuilder{
		res: &meshtimeout_api.MeshTimeoutResource{
			Meta: &test_model.ResourceMeta{
				Mesh:   core_model.DefaultMesh,
				Name:   "mt-1",
				Labels: map[string]string{},
			},
			Spec: &meshtimeout_api.MeshTimeout{},
		},
	}
}

func (m *MeshTimeoutBuilder) WithTargetRef(targetRef common_api.TargetRef) *MeshTimeoutBuilder {
	m.res.Spec.TargetRef = &targetRef
	return m
}

func (m *MeshTimeoutBuilder) WithName(name string) *MeshTimeoutBuilder {
	m.res.Meta.(*test_model.ResourceMeta).Name = name
	return m
}

func (m *MeshTimeoutBuilder) WithMesh(mesh string) *MeshTimeoutBuilder {
	m.res.Meta.(*test_model.ResourceMeta).Mesh = mesh
	return m
}

func (m *MeshTimeoutBuilder) AddFrom(targetRef common_api.TargetRef, conf meshtimeout_api.Conf) *MeshTimeoutBuilder {
	m.res.Spec.From = append(m.res.Spec.From, meshtimeout_api.From{
		TargetRef: targetRef,
		Default:   conf,
	})
	return m
}

func (m *MeshTimeoutBuilder) AddTo(targetRef common_api.TargetRef, conf meshtimeout_api.Conf) *MeshTimeoutBuilder {
	m.res.Spec.To = append(m.res.Spec.To, meshtimeout_api.To{
		TargetRef: targetRef,
		Default:   conf,
	})
	return m
}

func (m *MeshTimeoutBuilder) Build() *meshtimeout_api.MeshTimeoutResource {
	if err := m.res.Validate(); err != nil {
		panic(err)
	}
	return m.res
}
