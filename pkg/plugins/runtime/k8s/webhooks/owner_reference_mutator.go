package webhooks

import (
	"context"
	"encoding/json"
	"net/http"

	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_registry "github.com/kumahq/kuma/pkg/core/resources/registry"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	k8s_model "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

type OwnerReferenceMutator struct {
	Client       kube_client.Client
	CoreRegistry core_registry.TypeRegistry
	K8sRegistry  k8s_registry.TypeRegistry
	Decoder      *admission.Decoder
	Scheme       *kube_runtime.Scheme
}

func (m *OwnerReferenceMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
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
		// we need to also validate Mesh here because OwnerReferenceMutator is executed before validatingHandler
		if err := core_mesh.ValidateMesh(obj.GetMesh(), coreRes.Descriptor().Scope); err.HasViolations() {
			return convertValidationErrorOf(err, obj, obj.GetObjectMeta())
		}

		owner = &mesh_k8s.Mesh{}
		if err := m.Client.Get(ctx, kube_client.ObjectKey{Name: obj.GetMesh()}, owner); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
	}
	if err := controllerutil.SetOwnerReference(owner, obj, m.Scheme); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	mutatedRaw, err := json.Marshal(obj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, mutatedRaw)
}
