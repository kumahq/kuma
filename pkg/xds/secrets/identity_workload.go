package secrets

import (
	"github.com/pkg/errors"

	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	k8s_metadata "github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
)

// identityWorkload returns the workload used to derive a dataplane's mTLS
// identity (the SPIFFE SAN URI), keyed off the kuma.io/workload label — the
// same identity signal MeshService generation and the MeshIdentity
// Universal default already rely on.
func identityWorkload(dataplane *core_mesh.DataplaneResource) (string, error) {
	workload := dataplane.GetMeta().GetLabels()[k8s_metadata.KumaWorkload]
	if workload == "" {
		return "", errors.Errorf(
			"dataplane %q has no %s label, cannot derive mTLS identity",
			dataplane.GetMeta().GetName(), k8s_metadata.KumaWorkload,
		)
	}

	return workload, nil
}
