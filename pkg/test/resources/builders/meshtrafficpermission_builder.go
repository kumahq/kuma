package builders

import (
	"fmt"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	mtp_proto "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
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

func (m *MeshTrafficPermissionBuilder) AddRule(action mtp_proto.Action) *MeshTrafficPermissionBuilder {
	match := common_api.Match{
		SpiffeID: &common_api.SpiffeIDMatch{
			Type:  common_api.PrefixMatchType,
			Value: fmt.Sprintf("spiffe://%s", m.res.Meta.(*test_model.ResourceMeta).Mesh),
		},
	}
	var conf mtp_proto.RuleConf
	switch action {
	case mtp_proto.Deny:
		conf.Deny = pointer.To([]common_api.Match{match})
	case mtp_proto.AllowWithShadowDeny:
		conf.AllowWithShadowDeny = pointer.To([]common_api.Match{match})
	default:
		conf.Allow = pointer.To([]common_api.Match{match})
	}
	m.res.Spec.Rules = pointer.To(append(pointer.Deref(m.res.Spec.Rules), mtp_proto.Rule{Default: conf}))
	return m
}

func (m *MeshTrafficPermissionBuilder) Build() *mtp_proto.MeshTrafficPermissionResource {
	if err := m.res.Validate(); err != nil {
		panic(err)
	}
	return m.res
}
