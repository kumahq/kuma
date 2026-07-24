package secrets

import (
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	k8s_metadata "github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
)

// identityTags returns the tag set used to derive a dataplane's mTLS
// identity (SPIFFE/kuma SAN URIs). Dataplane.TagSet() is empty when
// Experimental.InboundTagsDisabled strips inbound tags, which would
// otherwise produce a certificate with no SAN and no Subject, rejected
// outright by Envoy. In that case, fall back to the kuma.io/workload
// label, the same identity signal MeshService generation and the
// MeshIdentity Universal default already rely on when tags are disabled.
func identityTags(dataplane *core_mesh.DataplaneResource) (mesh_proto.MultiValueTagSet, error) {
	tags := dataplane.Spec.TagSet()
	if len(tags.Values(mesh_proto.ServiceTag)) > 0 {
		return tags, nil
	}

	workload := dataplane.GetMeta().GetLabels()[k8s_metadata.KumaWorkload]
	if workload == "" {
		return nil, errors.Errorf(
			"dataplane %q has no %s tag and no %s label, cannot derive mTLS identity",
			dataplane.GetMeta().GetName(), mesh_proto.ServiceTag, k8s_metadata.KumaWorkload,
		)
	}

	return mesh_proto.MultiValueTagSetFrom(map[string][]string{
		mesh_proto.ServiceTag: {workload},
	}), nil
}
