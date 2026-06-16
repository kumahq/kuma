package labels

import (
	"maps"

	"github.com/pkg/errors"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v2/pkg/config/core"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
)

// Namespace type allows to avoid carrying both 'namespace' and 'systemNamespace' around the code base
// and depend on this type instead
type Namespace struct {
	value  string
	system bool
}

var UnsetNamespace = Namespace{}

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

type Options struct {
	Mode      config_core.CpMode
	IsK8s     bool
	ZoneName  string
	Namespace Namespace
	// Privileged marks trusted CP-internal writers. Non-local resources pass
	// through; local writes may keep OwnerSystem labels.
	Privileged bool
	// PreviousLabels preserves stored OwnerSystem labels on user updates.
	PreviousLabels map[string]string
}

type Option func(*Options)

func NewOptions(fs ...Option) *Options {
	opts := &Options{}
	for _, f := range fs {
		f(opts)
	}
	return opts
}

func WithK8s(k8s bool) Option {
	return func(opts *Options) {
		opts.IsK8s = k8s
	}
}

func WithNamespace(namespace Namespace) Option {
	return func(opts *Options) {
		opts.Namespace = namespace
	}
}

func WithZone(name string) Option {
	return func(opts *Options) {
		opts.ZoneName = name
	}
}

func WithMode(mode config_core.CpMode) Option {
	return func(opts *Options) {
		opts.Mode = mode
	}
}

func WithPrivileged(privileged bool) Option {
	return func(opts *Options) {
		opts.Privileged = privileged
	}
}

func WithPreviousLabels(labels map[string]string) Option {
	return func(opts *Options) {
		opts.PreviousLabels = labels
	}
}

// Compute returns the labels the CP should store for a resource.
func Compute(
	rd core_model.ResourceTypeDescriptor,
	spec core_model.ResourceSpec,
	existingLabels map[string]string,
	mesh string,
	displayName string,
	opts ...Option,
) (map[string]string, error) {
	o := NewOptions(opts...)
	labels := map[string]string{}
	if len(existingLabels) > 0 {
		labels = maps.Clone(existingLabels)
	}

	// Do not rewrite labels imported from another CP.
	if o.Privileged && !core_model.IsLocallyOriginated(o.Mode, labels) {
		return labels, nil
	}

	resourceMesh := mesh
	if rd.Scope == core_model.ScopeMesh && resourceMesh == "" {
		resourceMesh = core_model.DefaultMesh
	}

	env := config_core.UniversalEnvironment
	if o.IsK8s {
		env = config_core.KubernetesEnvironment
	}
	ctx := ValidationContext{
		Mode:         o.Mode,
		Env:          env,
		ZoneName:     o.ZoneName,
		Namespace:    o.Namespace,
		Descriptor:   rd,
		Spec:         spec,
		ResourceMesh: resourceMesh,
		ResourceName: displayName,
	}

	labels[mesh_proto.ResourceOriginLabel] = expectedOrigin(ctx)

	for key, specs := range registry {
		ls, ok := matchedSpec(specs, ctx)
		if !ok {
			delete(labels, key)
			continue
		}
		switch ls.Owner {
		case OwnerUser:
			continue
		case OwnerSystem:
			// Restore stored system labels on user writes; trust privileged writers.
			if !o.Privileged {
				if prev, ok := o.PreviousLabels[key]; ok {
					labels[key] = prev
				} else {
					delete(labels, key)
				}
			}
		case OwnerControlPlane:
			value, err := ls.Expected(ctx)
			if err != nil {
				return nil, err
			}
			labels[key] = value
		}
	}

	return labels, nil
}

func ComputePolicyRole(p core_model.Policy, ns Namespace) (mesh_proto.PolicyRole, error) {
	if ns.system || ns == UnsetNamespace {
		// on Universal the value is always empty
		return mesh_proto.SystemPolicyRole, nil
	}

	hasTo := false
	if pwtl, ok := p.(core_model.PolicyWithToList); ok && len(pwtl.GetToList()) > 0 {
		hasTo = true
	}

	hasFrom := false
	if pwfl, ok := p.(core_model.PolicyWithFromList); ok && len(pwfl.GetFromList()) > 0 {
		hasFrom = true
	}

	if hasFrom && hasTo {
		return "", errors.New("it's not allowed to mix 'to' and 'from' arrays in the same policy")
	}

	if hasFrom || (!hasTo && !hasFrom) {
		// if there is 'from' or neither (single item)
		return mesh_proto.WorkloadOwnerPolicyRole, nil
	}

	hasSameOrOmittedNamespace := func(tr common_api.TargetRef) bool {
		return pointer.Deref(tr.Namespace) == "" || pointer.Deref(tr.Namespace) == ns.value
	}

	isProducerItem := func(tr common_api.TargetRef) bool {
		switch tr.Kind {
		case common_api.MeshService, common_api.MeshHTTPRoute:
			return pointer.Deref(tr.Name) != "" && hasSameOrOmittedNamespace(tr)
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
