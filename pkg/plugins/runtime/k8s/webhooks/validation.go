package webhooks

import (
	"context"
	"fmt"
	"net/http"

	v1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_registry "github.com/kumahq/kuma/v3/pkg/core/resources/registry"
	"github.com/kumahq/kuma/v3/pkg/core/resources/validator"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
	k8s_common "github.com/kumahq/kuma/v3/pkg/plugins/common/k8s"
	mesh_k8s "github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s/native/api/v1alpha1"
	k8s_model "github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s/native/pkg/model"
	k8s_registry "github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s/native/pkg/registry"
)

func NewValidatingWebhook(
	converter k8s_common.Converter,
	coreRegistry core_registry.TypeRegistry,
	k8sRegistry k8s_registry.TypeRegistry,
	checker ResourceAdmissionChecker,
) k8s_common.AdmissionValidator {
	return &validatingHandler{
		coreRegistry:             coreRegistry,
		k8sRegistry:              k8sRegistry,
		converter:                converter,
		ResourceAdmissionChecker: checker,
	}
}

type validatingHandler struct {
	ResourceAdmissionChecker

	coreRegistry core_registry.TypeRegistry
	k8sRegistry  k8s_registry.TypeRegistry
	converter    k8s_common.Converter
	decoder      admission.Decoder
}

func (h *validatingHandler) InjectDecoder(d admission.Decoder) {
	h.decoder = d
}

func (h *validatingHandler) Handle(_ context.Context, req admission.Request) admission.Response {
	_, err := h.coreRegistry.DescriptorFor(core_model.ResourceType(req.Kind.Kind))
	if err != nil {
		// we only care about types in the registry for this handler
		return admission.Allowed("")
	}

	coreRes, k8sObj, err := h.decode(req)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if resp := h.IsOperationAllowed(req.UserInfo, coreRes, req.Namespace); !resp.Allowed {
		return resp
	}

	switch req.Operation {
	case v1.Delete:
		return admission.Allowed("")
	default:
		var warnings []string
		if err := core_mesh.ValidateMesh(k8sObj.GetMesh(), coreRes.Descriptor().Scope); err.HasViolations() {
			return convertValidationErrorOf(err, k8sObj, k8sObj.GetObjectMeta())
		}

		if err := h.validateLabels(coreRes.GetMeta()); err.HasViolations() {
			return convertValidationErrorOf(err, k8sObj, k8sObj.GetObjectMeta())
		}

		resp, err := h.validateOriginNotChanged(req, k8sObj)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		if resp != nil {
			return *resp
		}

		if !h.isPrivilegedUser(h.AllowedUsers, req.UserInfo) && coreRes.Descriptor().Scope == core_model.ScopeMesh {
			if err := h.validateMeshOwnerReference(k8sObj); err.HasViolations() {
				return convertValidationErrorOf(err, k8sObj, k8sObj.GetObjectMeta())
			}
		}

		if err := validator.Validate(coreRes); err != nil {
			if kumaErr, ok := err.(*validators.ValidationError); ok {
				// we assume that coreRes.Validate() returns validation errors of the spec
				return convertSpecValidationError(kumaErr, coreRes.Descriptor().IsPluginOriginated, k8sObj)
			}
			return admission.Denied(err.Error())
		}

		warnings = append(warnings, core_model.Deprecations(coreRes)...)
		return admission.Allowed("").WithWarnings(warnings...)
	}
}

func (h *validatingHandler) decode(req admission.Request) (core_model.Resource, k8s_model.KubernetesObject, error) {
	coreRes, err := h.coreRegistry.NewObject(core_model.ResourceType(req.Kind.Kind))
	if err != nil {
		return nil, nil, err
	}
	k8sObj, err := h.k8sRegistry.NewObject(coreRes.GetSpec())
	if err != nil {
		return nil, nil, err
	}

	switch req.Operation {
	case v1.Delete:
		if err := h.decoder.DecodeRaw(req.OldObject, k8sObj); err != nil {
			return nil, nil, err
		}
	default:
		if err := h.decoder.Decode(req, k8sObj); err != nil {
			return nil, nil, err
		}
	}

	if err := h.converter.ToCoreResource(k8sObj, coreRes); err != nil {
		return nil, nil, err
	}
	return coreRes, k8sObj, nil
}

// Without this a zone user could take over a Global-synced policy by re-applying it:
// the defaulting webhook recomputes kuma.io/origin to 'zone' for a non-privileged
// writer, and the Global->Zone KDS stream then wedges on AlreadyExists.
func (h *validatingHandler) validateOriginNotChanged(req admission.Request, newObj k8s_model.KubernetesObject) (*admission.Response, error) {
	if req.Operation != v1.Update {
		return nil, nil
	}
	// a non-federated zone owns everything in its store
	if h.Mode != core.Global && !h.FederatedZone {
		return nil, nil
	}
	// KDS sync and GC write on behalf of the control plane
	if h.isPrivilegedUser(h.AllowedUsers, req.UserInfo) {
		return nil, nil
	}

	oldObj, err := h.decodeOldObject(req)
	if err != nil {
		return nil, err
	}
	oldOrigin, ok := oldObj.GetLabels()[mesh_proto.ResourceOriginLabel]
	if !ok {
		return nil, nil
	}
	if newOrigin := newObj.GetLabels()[mesh_proto.ResourceOriginLabel]; newOrigin != oldOrigin {
		return forbiddenResponse(fmt.Sprintf(
			"Operation not allowed. '%s' label is immutable, cannot be changed from '%s' to '%s'",
			mesh_proto.ResourceOriginLabel, oldOrigin, newOrigin,
		)), nil
	}
	return nil, nil
}

func (h *validatingHandler) decodeOldObject(req admission.Request) (k8s_model.KubernetesObject, error) {
	coreRes, err := h.coreRegistry.NewObject(core_model.ResourceType(req.Kind.Kind))
	if err != nil {
		return nil, err
	}
	k8sObj, err := h.k8sRegistry.NewObject(coreRes.GetSpec())
	if err != nil {
		return nil, err
	}
	if err := h.decoder.DecodeRaw(req.OldObject, k8sObj); err != nil {
		return nil, err
	}
	return k8sObj, nil
}

func (h *validatingHandler) validateLabels(rm core_model.ResourceMeta) validators.ValidationError {
	var verr validators.ValidationError
	labelsPath := validators.Root().Field("labels")
	if origin, ok := core_model.ResourceOrigin(rm); ok {
		if err := origin.IsValid(); err != nil {
			verr.AddViolationAt(labelsPath.Key(mesh_proto.ResourceOriginLabel), err.Error())
		}
	}
	return verr
}

func (h *validatingHandler) validateMeshOwnerReference(obj k8s_model.KubernetesObject) validators.ValidationError {
	var verr validators.ValidationError
	path := validators.RootedAt("metadata").Field("ownerReferences")
	for i, ref := range obj.GetObjectMeta().GetOwnerReferences() {
		gv, err := schema.ParseGroupVersion(ref.APIVersion)
		if err != nil || gv.Group != mesh_k8s.GroupVersion.Group || ref.Kind != string(core_mesh.MeshType) {
			continue
		}
		if ref.Name != obj.GetMesh() {
			verr.AddViolationAt(path.Index(i).Field("name"), fmt.Sprintf(
				"must be the same as the mesh of the resource %q, got %q. A resource cannot be moved between meshes, delete it and apply it again in the new mesh", obj.GetMesh(), ref.Name,
			))
		}
	}
	return verr
}

func (h *validatingHandler) Supports(admission.Request) bool {
	return true
}

func convertSpecValidationError(kumaErr *validators.ValidationError, isPluginOriginated bool, obj k8s_model.KubernetesObject) admission.Response {
	verr := validators.OK()
	if kumaErr != nil {
		if isPluginOriginated {
			verr = *kumaErr
		} else {
			verr.AddError("spec", *kumaErr)
		}
	}
	return convertValidationErrorOf(verr, obj, obj.GetObjectMeta())
}

func convertValidationErrorOf(kumaErr validators.ValidationError, obj kube_runtime.Object, objMeta metav1.Object) admission.Response {
	details := &metav1.StatusDetails{
		Name: objMeta.GetName(),
		Kind: obj.GetObjectKind().GroupVersionKind().Kind,
	}
	resp := admission.Response{
		Allowed: false,
		Result: &metav1.Status{
			Status:  "Failure",
			Message: kumaErr.Error(),
			Reason:  "Invalid",
			Code:    int32(422),
			Details: details,
		},
	}
	for _, violation := range kumaErr.Violations {
		cause := metav1.StatusCause{
			Type:    "FieldValueInvalid",
			Message: violation.Message,
			Field:   violation.Field,
		}
		details.Causes = append(details.Causes, cause)
	}
	return resp
}
