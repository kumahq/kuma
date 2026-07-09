package labels

import (
	"maps"

	"github.com/pkg/errors"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
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
	Mode           config_core.CpMode
	IsK8s          bool
	ZoneName       string
	Namespace      Namespace
	ServiceAccount string
	Workload       string
	// Privileged marks trusted CP-internal writes (KDS sync, GC,
	// storage-version migrator) whose labels must not be recomputed.
	Privileged bool
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

func WithServiceAccount(name string) Option {
	return func(opts *Options) {
		opts.ServiceAccount = name
	}
}

func WithWorkload(name string) Option {
	return func(opts *Options) {
		opts.Workload = name
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

// Compute returns the labels the control plane should store for a resource. It
// force-sets / deletes control-plane-owned labels off the registry (see
// specs.go) and force-sets kuma.io/origin from the CP mode. User- and
// system-owned labels are left untouched.
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

	// Only skip recomputation for resources imported from another CP (e.g. via
	// KDS sync); locally-originated resources are always recomputed.
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
		ResourceName: displayName,
		ResourceMesh: resourceMesh,
	}

	labels[mesh_proto.ResourceOriginLabel] = expectedOrigin(ctx)

	for key, specs := range registry {
		ls, ok := matchedSpec(specs, ctx)
		switch {
		case ok && ls.Owner == OwnerControlPlane:
			value, err := ls.Expected(ctx)
			if err != nil {
				return nil, err
			}
			labels[key] = value
		case !ok && isControlPlaneKey(specs):
			delete(labels, key)
		}
	}

	// Service account and workload are not derived from the resource spec but
	// supplied by the caller (pod/gateway converters).
	if o.IsK8s && o.ServiceAccount != "" {
		labels[metadata.KumaServiceAccount] = o.ServiceAccount
	}
	if o.Workload != "" {
		labels[metadata.KumaWorkload] = o.Workload
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
