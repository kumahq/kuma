package dataplane

import (
	"context"
	"fmt"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

type membershipValidator struct {
}

var _ Validator = &membershipValidator{}

type NotAllowedErr struct {
	Mesh   string
	TagSet mesh_proto.SingleValueTagSet
}

func (n *NotAllowedErr) Error() string {
	return fmt.Sprintf("dataplane cannot be a member of mesh %q because its tags %q does not fulfill any requirements defined on the Mesh", n.Mesh, n.TagSet.String())
}

type DeniedErr struct {
	Mesh         string
	DpTagSet     mesh_proto.SingleValueTagSet
	DeniedTagSet mesh_proto.SingleValueTagSet
}

func (n *DeniedErr) Error() string {
	return fmt.Sprintf("dataplane cannot be a member of mesh %q because its tags %q matches restricted tags %q defined on the Mesh", n.Mesh, n.DpTagSet.String(), n.DeniedTagSet)
}

func NewMembershipValidator() Validator {
	return &membershipValidator{}
}

func (m *membershipValidator) ValidateCreate(_ context.Context, key model.ResourceKey, newDp *core_mesh.DataplaneResource, mesh *core_mesh.MeshResource) error {
	return m.validateDp(key, newDp, mesh)
}

func (m *membershipValidator) validateDp(key model.ResourceKey, dp *core_mesh.DataplaneResource, mesh *core_mesh.MeshResource) error {
	constraints := mesh.Spec.GetConstraints()
	if constraints == nil {
		return nil
	}
	dataplaneProxyConstraints := constraints.GetDataplaneProxy()
	if dataplaneProxyConstraints == nil {
		return nil
	}

	for _, tagSet := range dp.Spec.SingleValueTagSets() {
		if !isAllowedToJoin(tagSet, dataplaneProxyConstraints.GetRequirements()) {
			return &NotAllowedErr{
				Mesh:   key.Mesh,
				TagSet: tagSet,
			}
		}
		if denied, deniedTags := isDeniedToJoin(tagSet, dataplaneProxyConstraints.GetRestrictions()); denied {
			return &DeniedErr{
				Mesh:         key.Mesh,
				DpTagSet:     tagSet,
				DeniedTagSet: deniedTags,
			}
		}
	}
	return nil
}

func isAllowedToJoin(tagSet mesh_proto.SingleValueTagSet, requirements []*mesh_proto.Mesh_DataplaneProxyConstraints_Rules) bool {
	if len(requirements) == 0 {
		return true
	}
	for _, requirement := range requirements {
		if matchTags(requirement.Tags, tagSet) {
			return true
		}
	}
	return false
}

func isDeniedToJoin(tagSet mesh_proto.SingleValueTagSet, restrictions []*mesh_proto.Mesh_DataplaneProxyConstraints_Rules) (bool, mesh_proto.SingleValueTagSet) {
	for _, restriction := range restrictions {
		if matchTags(restriction.Tags, tagSet) {
			return true, restriction.Tags
		}
	}
	return false, nil
}

// matchTags checks whether dp tags has required tags
// If required tags has `*` in value, dp tags have to contain non-empty value
func matchTags(requiredTags map[string]string, dpTags map[string]string) bool {
	for requiredTag, requiredValue := range requiredTags {
		dpValue := dpTags[requiredTag]
		if requiredValue == mesh_proto.MatchAllTag && dpValue != "" {
			continue
		}
		if requiredValue != dpValue {
			return false
		}
	}
	return true
}

func (m *membershipValidator) ValidateUpdate(_ context.Context, newDp *core_mesh.DataplaneResource, mesh *core_mesh.MeshResource) error {
	return m.validateDp(model.MetaToResourceKey(newDp.GetMeta()), newDp, mesh)
}
