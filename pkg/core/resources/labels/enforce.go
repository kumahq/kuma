package labels

import (
	kuma_cp "github.com/kumahq/kuma/v3/pkg/config/app/kuma-cp"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

// ControlPlane is the configuration fixed for the whole deployment. Its zero value is
// for callers that must not enforce origin and zone, such as the admission webhooks,
// which validate the user-supplied labels instead.
type ControlPlane struct {
	Mode          config_core.CpMode
	Zone          string
	IsK8s         bool
	FederatedZone bool
}

func ControlPlaneFromConfig(cfg kuma_cp.Config) ControlPlane {
	return ControlPlane{
		Mode:          cfg.Mode,
		Zone:          cfg.Multizone.Zone.Name,
		IsK8s:         cfg.Environment == config_core.KubernetesEnvironment,
		FederatedZone: cfg.IsFederatedZoneCP(),
	}
}

// Write is the input to Compute and Validate: a resource as a writer submits it. Mesh
// and DisplayName come from the request rather than from the Meta.
type Write struct {
	Descriptor  core_model.ResourceTypeDescriptor
	Spec        core_model.ResourceSpec
	Namespace   Namespace
	Mesh        string
	DisplayName string
	// Labels as submitted. Read-only.
	Labels map[string]string
	// Previous: the labels stored for the object an update replaces; nil on a create.
	// A submitted value equal to the stored one is the object round-tripping through
	// the writer, not a value the writer chose, so ValidateOwnership skips it.
	Previous map[string]string
	// TrustedWriter: the labels come from a control plane (the store, KDS, GC, the
	// storage-version migrator, the CP's own k8s controllers), not from a user.
	TrustedWriter bool
	// Only the pod converter sets these.
	ServiceAccount string
	Workload       string
}

// StoredResource is the input to EnforcedReadLabels. It takes Descriptor and Spec
// apart from core_model.Resource because the resource's Meta is not set yet.
type StoredResource struct {
	Descriptor core_model.ResourceTypeDescriptor
	Spec       core_model.ResourceSpec
	Namespace  Namespace
	IsLocal    bool
}

// NewStoredResource derives IsLocal: KDS only writes into the system namespace, so
// anything outside it is local; inside it, and on Universal, the stored origin is trusted
// because the API server recomputes it on every write and the CP is the only other writer.
func NewStoredResource(res core_model.Resource, ns Namespace, storedLabels map[string]string, cp ControlPlane) StoredResource {
	return StoredResource{
		Descriptor: res.Descriptor(),
		Spec:       res.GetSpec(),
		Namespace:  ns,
		IsLocal:    (ns.value != "" && !ns.system) || core_model.IsLocallyOriginated(cp.Mode, storedLabels),
	}
}

// EnforcedReadLabels recomputes the control-plane-owned labels from the object's own
// identity, never from storage: a stored object can predate a label or have been
// written while the defaulting webhook did not cover its namespace.
func EnforcedReadLabels(r StoredResource, cp ControlPlane) map[string]string {
	var enforced map[string]string
	for _, d := range registry {
		if d.EnforceOnRead == nil {
			continue
		}
		v, ok := d.EnforceOnRead(r, cp)
		if !ok {
			continue
		}
		if enforced == nil {
			enforced = map[string]string{}
		}
		enforced[d.Key] = v
	}
	return enforced
}
