package webhooks

import (
	"context"
	"encoding/json"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_registry "github.com/kumahq/kuma/v3/pkg/core/resources/registry"
	mesh_k8s "github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s/native/api/v1alpha1"
	k8s_model "github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s/native/pkg/model"
	k8s_registry "github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s/native/pkg/registry"
)

// GarbageCollectorUser is the Kubernetes garbage collector. It strips owner references from
// dependents when an owner is deleted with the orphan propagation policy, so we must neither
// add them back nor reject its requests.
const GarbageCollectorUser = "system:serviceaccount:kube-system:generic-garbage-collector"

type OwnerReferenceMutator struct {
	Client                 kube_client.Client
	CoreRegistry           core_registry.TypeRegistry
	K8sRegistry            k8s_registry.TypeRegistry
	Decoder                admission.Decoder
	Scheme                 *kube_runtime.Scheme
	SkipMeshOwnerReference bool
	CpMode                 config_core.CpMode
}

func (m *OwnerReferenceMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	if req.UserInfo.Username == GarbageCollectorUser {
		return admission.Allowed("ignored. Request from the Kubernetes garbage collector.")
	}

	resType := core_model.ResourceType(req.Kind.Kind)

	coreRes, err := m.CoreRegistry.NewObject(resType)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	obj, err := m.K8sRegistry.NewObject(coreRes.GetSpec())
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// unmarshal k8s object from the request
	if err := m.Decoder.Decode(req, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	var owner k8s_model.KubernetesObject
	switch resType {
	case meshservice_api.MeshServiceType:
		return admission.Allowed("ignored. MeshService has a reference for Service")
	case core_mesh.DataplaneInsightType:
		owner = &mesh_k8s.Dataplane{}
		if err := m.Client.Get(ctx, kube_client.ObjectKey{Name: obj.GetName(), Namespace: obj.GetNamespace()}, owner); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
	default:
		if m.SkipMeshOwnerReference {
			return admission.Allowed("ignored. Configuration setup to ignore Mesh owner reference.")
		}
		if coreRes.Descriptor().Scope != core_model.ScopeMesh {
			return admission.Allowed("ignored. It's not a Mesh scoped resource.")
		}
		// we need to also validate Mesh here because OwnerReferenceMutator is executed before validatingHandler
		if err := core_mesh.ValidateMesh(obj.GetMesh(), coreRes.Descriptor().Scope); err.HasViolations() {
			return convertValidationErrorOf(err, obj, obj.GetObjectMeta())
		}

		if syncedResource(m.CpMode, obj.GetLabels()) {
			return admission.Allowed("ignore. It's synced resource.")
		}

		// the resource could have been moved to another Mesh, in which case a reference to the
		// previous Mesh has to go away. Otherwise deleting the previous Mesh garbage collects
		// a resource that belongs to a different Mesh.
		removeStaleMeshOwnerReferences(obj)

		owner = &mesh_k8s.Mesh{}
		if err := m.Client.Get(ctx, kube_client.ObjectKey{Name: obj.GetMesh()}, owner); err != nil {
			if req.Operation == admissionv1.Update && kube_apierrs.IsNotFound(err) {
				// the Mesh is already gone, don't block updates of the resource. Stale
				// references are dropped anyway.
				return patchResponse(req, obj)
			}
			return admission.Errored(http.StatusBadRequest, err)
		}
	}
	// owner references are matched by name, so this also refreshes the UID of a Mesh
	// that was deleted and created again.
	if err := controllerutil.SetOwnerReference(owner, obj, m.Scheme); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	return patchResponse(req, obj)
}

func patchResponse(req admission.Request, obj k8s_model.KubernetesObject) admission.Response {
	mutatedRaw, err := json.Marshal(obj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, mutatedRaw)
}

// removeStaleMeshOwnerReferences drops references to Meshes other than the one the resource
// belongs to. References to other kinds of owners, like the Service of a MeshZoneAddress,
// are preserved.
func removeStaleMeshOwnerReferences(obj k8s_model.KubernetesObject) {
	refs := obj.GetOwnerReferences()
	kept := make([]kube_meta.OwnerReference, 0, len(refs))
	for _, ref := range refs {
		if isMeshOwnerReference(ref) && ref.Name != obj.GetMesh() {
			continue
		}
		kept = append(kept, ref)
	}
	if len(kept) != len(refs) {
		obj.SetOwnerReferences(kept)
	}
}

func isMeshOwnerReference(ref kube_meta.OwnerReference) bool {
	gv, err := schema.ParseGroupVersion(ref.APIVersion)
	if err != nil {
		return false
	}
	return gv.Group == mesh_k8s.GroupVersion.Group && ref.Kind == string(core_mesh.MeshType)
}

func syncedResource(cpMode config_core.CpMode, labels map[string]string) bool {
	syncedOrigin := mesh_proto.GlobalResourceOrigin
	if cpMode == config_core.Global {
		syncedOrigin = mesh_proto.ZoneResourceOrigin
	}
	return len(labels) > 0 && labels[mesh_proto.ResourceOriginLabel] == string(syncedOrigin)
}
