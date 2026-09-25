package labels

import (
	"maps"

	"github.com/pkg/errors"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

// Namespace type allows to avoid carrying both 'namespace' and 'systemNamespace' around the code base
// and depend on this type instead
type Namespace struct {
	value  string
	system bool
}

var UnsetNamespace = Namespace{}

// Labels the control plane used to compute and no longer does. They are
// deleted on every proxy write so a resource created by an older control plane
// stops carrying them, instead of keeping a value nothing maintains.
var removedLabels = []string{"kuma.io/proxy-type", "kuma.io/gateway"}

func NewNamespace(value string, system bool) Namespace {
	return Namespace{
		value:  value,
		system: system,
	}
}

func GetNamespace(rm core_model.ResourceMeta, systemNamespace string) Namespace {
	if ns, ok := rm.GetNameExtensions()[core_model.K8sNamespaceComponent]; ok && ns != "" {
		return Namespace{
			value:  ns,
			system: ns == systemNamespace,
		}
	}
	return UnsetNamespace
}

// Compute returns the labels to store on a write: the supplied labels with every
// registered rule applied. A trusted write of a resource this control plane does not
// own (an import synced by KDS) is stored as is.
func Compute(w Write, cp ControlPlane) (map[string]string, error) {
	labels := map[string]string{}
	maps.Copy(labels, w.Labels)
	if w.TrustedWriter && !core_model.IsLocallyOriginated(cp.Mode, labels) {
		return labels, nil
	}
	for _, d := range registry {
		if d.Compute == nil {
			continue
		}
		v, ok, err := d.Compute(d.Key, w, cp)
		if err != nil {
			return nil, err
		}
		if ok {
			labels[d.Key] = v
		} else {
			delete(labels, d.Key)
		}
	}
	if w.Descriptor.IsProxy {
		for _, k := range removedLabels {
			delete(labels, k)
		}
	}
	return labels, nil
}

// ComputePolicyRole classifies a policy from its own placement and its to[] items.
// zone is the policy's own kuma.io/zone; it is empty for a policy that does not
// originate from a zone.
func ComputePolicyRole(p core_model.Policy, ns Namespace, zone string) (mesh_proto.PolicyRole, error) {
	if ns.system || ns == UnsetNamespace {
		// on Universal the value is always empty
		return mesh_proto.SystemPolicyRole, nil
	}

	hasTo := false
	if pwtl, ok := p.(core_model.PolicyWithToList); ok && len(pwtl.GetToList()) > 0 {
		hasTo = true
	}

	if !hasTo {
		// single-item and rules-based inbound policies remain workload-owner scoped
		return mesh_proto.WorkloadOwnerPolicyRole, nil
	}

	// selectsOwnResource reports whether ref names a single resource the policy's
	// own namespace owns. A label selector is resolved as a subset match over the
	// whole mesh, so only display-name, namespace and zone together pin it to one
	// resource. Accepting fewer labels would let a producer policy, which is
	// applied to every dataplane in the mesh and synced to the other zones, attach
	// to a resource of another namespace or zone.
	selectsOwnResource := func(tr common_api.TargetRef) bool {
		labels := pointer.Deref(tr.Labels)
		if len(labels) != 3 || zone == "" {
			return false
		}
		return labels[mesh_proto.DisplayName] != "" &&
			labels[mesh_proto.KubeNamespaceTag] == ns.value &&
			labels[mesh_proto.ZoneTag] == zone
	}

	isProducerItem := func(tr common_api.TargetRef) bool {
		switch tr.Kind {
		case common_api.MeshService, common_api.MeshHTTPRoute:
			return selectsOwnResource(tr)
		default:
			return false
		}
	}

	producerItems := 0
	for _, item := range p.(core_model.PolicyWithToList).GetToList() {
		if isProducerItem(item.GetTargetRef()) {
			producerItems++
		}
	}

	switch {
	case producerItems == len(p.(core_model.PolicyWithToList).GetToList()):
		return mesh_proto.ProducerPolicyRole, nil
	case producerItems == 0:
		return mesh_proto.ConsumerPolicyRole, nil
	default:
		return "", errors.New("it's not allowed to mix producer and consumer items in the same policy")
	}
}
