package builders

import (
	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	"github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	meshtimeout_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
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
	m.res.Spec.TargetRef = pointer.To(ToTopLevelTargetRef(targetRef))
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

func (m *MeshTimeoutBuilder) WithNamespace(namespace string) *MeshTimeoutBuilder {
	if m.res.Meta.(*test_model.ResourceMeta).NameExtensions == nil {
		m.res.Meta.(*test_model.ResourceMeta).NameExtensions = core_model.ResourceNameExtensions{}
	}
	m.res.Meta.(*test_model.ResourceMeta).NameExtensions[v1alpha1.KubeNamespaceTag] = namespace
	return m
}

func (m *MeshTimeoutBuilder) AddTo(targetRef common_api.TargetRef, conf meshtimeout_api.Conf) *MeshTimeoutBuilder {
	m.res.Spec.To = pointer.To(append(pointer.Deref(m.res.Spec.To), meshtimeout_api.To{
		TargetRef: ToOutboundTargetRef(targetRef),
		Default:   conf,
	}))
	return m
}

func (m *MeshTimeoutBuilder) AddRule(matches []common_api.Match, conf meshtimeout_api.Conf) *MeshTimeoutBuilder {
	rule := meshtimeout_api.Rule{Default: conf}
	if len(matches) > 0 {
		rule.Matches = pointer.To(matches)
	}
	m.res.Spec.Rules = pointer.To(append(pointer.Deref(m.res.Spec.Rules), rule))
	return m
}

func (m *MeshTimeoutBuilder) Build() *meshtimeout_api.MeshTimeoutResource {
	if err := m.res.Validate(); err != nil {
		panic(err)
	}
	return m.res
}
