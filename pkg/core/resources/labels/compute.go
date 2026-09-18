package labels

import (
	"maps"

	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

// Namespace type allows to avoid carrying both 'namespace' and 'systemNamespace' around the code base
// and depend on this type instead
type Namespace struct {
	value  string
	system bool
}

var UnsetNamespace = Namespace{}

// Labels the control plane used to compute and no longer does. They are deleted on
// every write of the kinds they were computed for, so a resource created by an older
// control plane stops carrying them, instead of keeping a value nothing maintains.
var (
	removedProxyLabels  = []string{"kuma.io/proxy-type", "kuma.io/gateway"}
	removedPolicyLabels = []string{"kuma.io/policy-role"}
)

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
		v, ok, err := d.Compute(w, cp)
		if err != nil {
			return nil, err
		}
		if ok {
			labels[d.Key] = v
		} else {
			delete(labels, d.Key)
		}
	}
	var removed []string
	switch {
	case w.Descriptor.IsProxy:
		removed = removedProxyLabels
	case w.Descriptor.IsPolicy:
		removed = removedPolicyLabels
	}
	for _, k := range removed {
		delete(labels, k)
	}
	return labels, nil
}
