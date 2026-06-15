package labels

import (
	"maps"

	"github.com/pkg/errors"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v2/pkg/config/core"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
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
	Mode           config_core.CpMode
	IsK8s          bool
	ZoneName       string
	Namespace      Namespace
	ServiceAccount string
	Workload       string
	// Privileged marks the caller as a trusted CP-internal flow (KDS sync,
	// GC, storage-version migrator, K8s controllers writing under the CP
	// service account). It changes Compute's behavior in two ways:
	//
	//   - If the resource is not locally originated (e.g. kuma.io/origin
	//     reports a different CP), Compute returns the labels unchanged.
	//     KDS-synced resources are authoritative from the source CP and
	//     must not be rewritten.
	//   - OwnerSystem labels the caller supplied are preserved instead of
	//     stripped. K8s controllers legitimately set keys like
	//     kuma.io/managed-by; user input paths still drop them.
	Privileged bool
	// PreviousLabels are the labels currently persisted for the resource.
	// On non-Privileged user paths, Compute restores OwnerSystem keys from
	// here instead of dropping them — controllers like the Universal
	// workload-generator stamp kuma.io/managed-by directly via the store,
	// and a subsequent user PUT must not wipe it out.
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

func WithPreviousLabels(labels map[string]string) Option {
	return func(opts *Options) {
		opts.PreviousLabels = labels
	}
}

// Compute returns the label map the CP would store for the given resource.
//
// If Options.Privileged is set and the resource is not locally originated
// (the KDS-sync case — labels are authoritative from another CP), Compute
// returns the labels untouched.
//
// Otherwise it walks the LabelSpec registry and, for each non-user-owned key:
//
//   - OwnerControlPlane, Expected applies=true, !OpenValue: force-set the
//     expected value, overriding whatever the user supplied.
//   - OwnerControlPlane, Expected applies=true, OpenValue: keep the user
//     value as-is. The CP doesn't dictate the value, just where it lives.
//   - OwnerControlPlane, Expected applies=false: delete the key. The label
//     isn't applicable in this context; a user value would be stale.
//   - OwnerSystem: on Privileged paths, preserve the caller's value. On
//     user paths, restore from PreviousLabels if it carries a value for
//     this key (the previously-stored CP-set value is authoritative);
//     otherwise drop. User input on these keys is never trusted —
//     controllers like the Universal workload-generator stamp
//     kuma.io/managed-by directly via the store, and a subsequent user
//     PUT must not wipe it out.
//
// OwnerUser keys are left untouched.
//
// After the registry walk, Compute writes the OwnerSystem labels it owns
// based on caller-supplied options and the resource spec (ServiceAccount,
// Workload, and the Dataplane listener flags).
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

	// Privileged callers writing a non-locally-originated resource (the KDS
	// sync case) bring labels that are authoritative from the source CP —
	// rewriting them here would clobber kuma.io/origin, kuma.io/zone, etc.
	// Pass them through unchanged.
	if o.Privileged && !core_model.IsLocallyOriginated(o.Mode, labels) {
		return labels, nil
	}

	resourceMesh := mesh
	if rd.Scope == core_model.ScopeMesh && resourceMesh == "" {
		resourceMesh = core_model.DefaultMesh
	}

	ctx := ValidationContext{
		Mode:         o.Mode,
		IsK8s:        o.IsK8s,
		ZoneName:     o.ZoneName,
		Namespace:    o.Namespace,
		Descriptor:   rd,
		Spec:         spec,
		ResourceMesh: resourceMesh,
		ResourceName: displayName,
	}

	// kuma.io/origin sits outside the registry. Force-set the CP-computed
	// value — origin is deterministic from the CP mode.
	labels[mesh_proto.ResourceOriginLabel] = expectedOrigin(ctx)

	for _, ls := range registry {
		switch ls.Owner {
		case OwnerUser:
			continue
		case OwnerSystem:
			// Privileged callers (K8s controllers, etc.) legitimately set
			// OwnerSystem labels like kuma.io/managed-by. On user paths the
			// user's own value is never trusted, but the previously-stored
			// CP value must survive — the Universal workload-generator and
			// meshservice-generator stamp these directly via the store, and
			// a subsequent user PUT would otherwise drop them.
			if !o.Privileged {
				if prev, ok := o.PreviousLabels[ls.Key]; ok {
					labels[ls.Key] = prev
				} else {
					delete(labels, ls.Key)
				}
			}
		case OwnerControlPlane:
			value, applies := "", true
			if ls.Expected != nil {
				value, applies = ls.Expected(ctx)
			}
			if !applies {
				delete(labels, ls.Key)
				continue
			}
			if ls.OpenValue {
				continue
			}
			labels[ls.Key] = value
		}
	}

	// ComputePolicyRole reports two spec-validation errors (from+to mixed,
	// producer+consumer mixed) that the registry's Expected swallows as
	// applies=false. Surface them explicitly here so the create/update fails
	// rather than silently dropping kuma.io/policy-role.
	if rd.IsPolicy && rd.IsPluginOriginated && o.Namespace.value != "" {
		if _, err := ComputePolicyRole(spec.(core_model.Policy), o.Namespace); err != nil {
			return nil, err
		}
	}

	if dp, ok := spec.(*mesh_proto.Dataplane); ok {
		for _, l := range dp.GetNetworking().GetListeners() {
			switch l.Type {
			case mesh_proto.Dataplane_Networking_Listener_ZoneIngress:
				labels[mesh_proto.ListenerZoneIngressLabel] = "enabled"
			case mesh_proto.Dataplane_Networking_Listener_ZoneEgress:
				labels[mesh_proto.ListenerZoneEgressLabel] = "enabled"
			}
		}
	}

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
