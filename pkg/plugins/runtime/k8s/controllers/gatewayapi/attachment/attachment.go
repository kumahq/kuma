package attachment

import (
	"context"

	kube_core "k8s.io/api/core/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"
)

type Attachment int

const (
	// Unknown means we shouldn't report on whether it's attached or not
	Unknown Attachment = iota
	// Allowed means we could successfully attach
	Allowed
)

type Kind int

const (
	UnknownKind Kind = iota
	Service
)

// EvaluateParentRefAttachment reports whether a route in the given namespace can attach
// via the given ParentRef.
func EvaluateParentRefAttachment(
	ctx context.Context,
	client kube_client.Client,
	routeHostnames []gatewayapi.Hostname,
	routeNs *kube_core.Namespace,
	ref gatewayapi.ParentReference,
) (Attachment, Kind, error) {
	if *ref.Kind == "Service" && (*ref.Group == kube_core.GroupName || *ref.Group == gatewayapi.GroupName) {
		// Attaching to a Service can only affect requests coming from Services
		// in the same Namespace or requests going to the Service.
		return Allowed, Service, nil
	}
	return Unknown, UnknownKind, nil
}
