package labels

import (
	"fmt"
	"maps"
	"strings"

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

// Compute computes labels for a resource based on its type, spec, existing labels, namespace, mesh, mode, k8s and localZone.
// Only use set / setIfNotExist to set labels as it makes sure the label is on the list of computed labels (that is used in another project).
func Compute(
	rd core_model.ResourceTypeDescriptor,
	spec core_model.ResourceSpec,
	existingLabels map[string]string,
	mesh string,
	opts ...Option,
) (map[string]string, error) {
	labelsOpts := NewOptions(opts...)
	labels := map[string]string{}
	if len(existingLabels) > 0 {
		labels = maps.Clone(existingLabels)
	}

	set := func(k, v string) {
		if _, ok := AllComputedLabels[k]; !ok {
			panic(fmt.Sprintf("label %q is not in the list of computed labels, update AllComputedLabels list as it is used in another project", k))
		}
		labels[k] = v
	}

	setIfNotExist := func(k, v string) {
		if _, ok := labels[k]; !ok {
			set(k, v)
		}
	}

	getMeshOrDefault := func() string {
		if mesh != "" {
			return mesh
		}
		return core_model.DefaultMesh
	}

	if rd.Scope == core_model.ScopeMesh {
		setIfNotExist(metadata.KumaMeshLabel, getMeshOrDefault())
	}

	if labelsOpts.Mode == config_core.Zone {
		// If resource can't be created on Zone (like Mesh), there is no point in adding
		// 'kuma.io/zone', 'kuma.io/origin' and 'kuma.io/env' labels even if the zone is non-federated
		if rd.KDSFlags.Has(core_model.ProvidedByZoneFlag) {
			setIfNotExist(mesh_proto.ResourceOriginLabel, string(mesh_proto.ZoneResourceOrigin))
			if labels[mesh_proto.ResourceOriginLabel] != string(mesh_proto.GlobalResourceOrigin) {
				setIfNotExist(mesh_proto.ZoneTag, labelsOpts.ZoneName)
				env := mesh_proto.UniversalEnvironment
				if labelsOpts.IsK8s {
					env = mesh_proto.KubernetesEnvironment
				}
				setIfNotExist(mesh_proto.EnvTag, env)
			}
		}
	}

	if labelsOpts.Namespace.value != "" && labelsOpts.IsK8s && core_model.IsLocallyOriginated(labelsOpts.Mode, labels) {
		setIfNotExist(mesh_proto.KubeNamespaceTag, labelsOpts.Namespace.value)
	}

	if labelsOpts.Namespace.value != "" && rd.IsPolicy && rd.IsPluginOriginated && core_model.IsLocallyOriginated(labelsOpts.Mode, labels) {
		role, err := ComputePolicyRole(spec.(core_model.Policy), labelsOpts.Namespace)
		if err != nil {
			return nil, err
		}
		set(mesh_proto.PolicyRoleLabel, string(role))
	}

	if rd.IsProxy {
		proxy, ok := spec.(core_model.ProxyResource)
		if ok {
			set(mesh_proto.ProxyTypeLabel, strings.ToLower(string(proxy.GetProxyType())))
		}
	}

	if labelsOpts.IsK8s {
		if labelsOpts.ServiceAccount != "" {
			set(metadata.KumaServiceAccount, labelsOpts.ServiceAccount)
		}
	}

	if labelsOpts.Workload != "" {
		set(metadata.KumaWorkload, labelsOpts.Workload)
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
