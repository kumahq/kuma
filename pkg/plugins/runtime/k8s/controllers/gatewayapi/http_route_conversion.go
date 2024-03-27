package gatewayapi

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_apimeta "k8s.io/apimachinery/pkg/api/meta"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/metadata"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi/attachment"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi/referencegrants"
	k8s_util "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
)

type ResolvedRefsConditionFalse struct {
	Reason  string
	Message string
}

func (c *ResolvedRefsConditionFalse) AddIfFalseAndNotPresent(conditions *[]kube_meta.Condition) {
	if c != nil && kube_apimeta.FindStatusCondition(*conditions, string(gatewayapi.RouteConditionResolvedRefs)) == nil {
		condition := kube_meta.Condition{
			Type:    string(gatewayapi.RouteReasonResolvedRefs),
			Status:  kube_meta.ConditionFalse,
			Reason:  c.Reason,
			Message: c.Message,
		}
		kube_apimeta.SetStatusCondition(conditions, condition)
	}
}

func (r *HTTPRouteReconciler) uncheckedGapiToKumaRef(
	ctx context.Context, mesh string, objectNamespace string, ref gatewayapi.BackendObjectReference,
) (map[string]string, *ResolvedRefsConditionFalse, error) {
	unresolvedBackendTags := map[string]string{
		mesh_proto.ServiceTag: metadata.UnresolvedBackendServiceTag,
	}

	policyRef := referencegrants.PolicyReferenceBackend(referencegrants.FromHTTPRouteIn(objectNamespace), ref)

	gk := policyRef.GroupKindReferredTo()
	namespacedName := policyRef.NamespacedNameReferredTo()

	switch {
	case gk.Kind == "Service" && gk.Group == "":
		// References to Services are required by GAPI to include a port
		port := int32(*ref.Port)

		svc := &kube_core.Service{}
		if err := r.Client.Get(ctx, namespacedName, svc); err != nil {
			if kube_apierrs.IsNotFound(err) {
				return unresolvedBackendTags,
					&ResolvedRefsConditionFalse{
						Reason:  string(gatewayapi.RouteReasonBackendNotFound),
						Message: fmt.Sprintf("backend reference references a non-existent Service %q", namespacedName.String()),
					},
					nil
			}
			return nil, nil, err
		}

		return map[string]string{
			mesh_proto.ServiceTag: k8s_util.ServiceTag(kube_client.ObjectKeyFromObject(svc), &port),
		}, nil, nil
	case gk.Kind == "ExternalService" && gk.Group == mesh_k8s.GroupVersion.Group:
		resource := core_mesh.NewExternalServiceResource()
		if err := r.ResourceManager.Get(ctx, resource, store.GetByKey(namespacedName.Name, mesh)); err != nil {
			if store.IsResourceNotFound(err) {
				return unresolvedBackendTags,
					&ResolvedRefsConditionFalse{
						Reason:  string(gatewayapi.RouteReasonBackendNotFound),
						Message: fmt.Sprintf("backend reference references a non-existent ExternalService %q", namespacedName.Name),
					},
					nil
			}
			return nil, nil, err
		}

		return map[string]string{
			mesh_proto.ServiceTag: resource.Spec.GetService(),
		}, nil, nil
	}

	return unresolvedBackendTags,
		&ResolvedRefsConditionFalse{
			Reason:  string(gatewayapi.RouteReasonInvalidKind),
			Message: "backend reference must be Service or externalservice.kuma.io",
		},
		nil
}

// gapiToKumaRef checks a reference and tries to resolve if it's supported by
// Kuma. It returns a condition with Reason/Message if it fails or an error for
// unexpected errors.
func (r *HTTPRouteReconciler) gapiToKumaRef(
	ctx context.Context,
	mesh string,
	objectNamespace string,
	ref gatewayapi.BackendObjectReference,
	refAttachmentKind attachment.Kind,
) (map[string]string, *ResolvedRefsConditionFalse, error) {
	// ReferenceGrants don't need to be taken into account for Mesh
	if refAttachmentKind != attachment.Service {
		unresolvedBackendTags := map[string]string{
			mesh_proto.ServiceTag: metadata.UnresolvedBackendServiceTag,
		}

		policyRef := referencegrants.PolicyReferenceBackend(referencegrants.FromHTTPRouteIn(objectNamespace), ref)

		gk := policyRef.GroupKindReferredTo()
		namespacedName := policyRef.NamespacedNameReferredTo()

		if permitted, err := referencegrants.IsReferencePermitted(ctx, r.Client, policyRef); err != nil {
			return nil, nil, errors.Wrap(err, "couldn't determine if backend reference is permitted")
		} else if !permitted {
			return unresolvedBackendTags,
				&ResolvedRefsConditionFalse{
					Reason:  string(gatewayapi.RouteReasonRefNotPermitted),
					Message: fmt.Sprintf("reference to %s %q not permitted by any ReferenceGrant", gk, namespacedName),
				},
				nil
		}
	}

	return r.uncheckedGapiToKumaRef(ctx, mesh, objectNamespace, ref)
}
