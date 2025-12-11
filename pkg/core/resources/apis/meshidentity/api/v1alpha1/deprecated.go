package v1alpha1

import (
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
)

func deprecations(r *model.ResStatus[*MeshIdentity, *MeshIdentityStatus]) []string {
	return nil
}
