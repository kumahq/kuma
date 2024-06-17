package webhooks

import (
	"context"
	"encoding/json"
	"net/http"

	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
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
	CpMode       config_core.CpMode
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

		if syncedResource(m.CpMode, obj.GetLabels()) {
			return admission.Allowed("ignore. It's synced resource.")
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

func syncedResource(cpMode config_core.CpMode, labels map[string]string) bool {
	syncedOrigin := mesh_proto.GlobalResourceOrigin
	if cpMode == config_core.Global {
		syncedOrigin = mesh_proto.ZoneResourceOrigin
	}
	return len(labels) > 0 && labels[mesh_proto.ResourceOriginLabel] == string(syncedOrigin)
}
