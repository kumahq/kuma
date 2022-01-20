package attachment

import (
	"context"

	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_labels "k8s.io/apimachinery/pkg/labels"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi/common"
)

type Attachment int

const (
	// Unknown means we shouldn't report on whether it's attached or not
	Unknown Attachment = iota
	// NotPermitted means the route isn't allowed to attach to the ref
	NotPermitted
	// Invalid means the route points to a nonexistent parent
	// or is otherwise invalid
	Invalid
	// Allowed means we could successfully attach
	Allowed
)

// findRouteListenerAttachment reports whether this ref is allowed to attach to
// the Gateway or specific Listener.
// refSectionName is nil if the ref refers to the whole Gateway.
func findRouteListenerAttachment(
	gateway *gatewayapi.Gateway,
	routeNs kube_client.Object,
	refSectionName *gatewayapi.SectionName,
) (Attachment, error) {
	// Build a map of whether attaching to each listener is possible
	listeners := map[gatewayapi.SectionName]Attachment{}

	for _, l := range gateway.Spec.Listeners {
		ns := l.AllowedRoutes.Namespaces

		// The gateway controller ensures Kinds is either empty or contains
		// HTTPRoute, so we can ignore it here.

		// From determines whether we are permitted to attach to this ParentRef
		switch *ns.From {
		case gatewayapi.NamespacesFromSelector:
			// TODO, the gateway controller/webhook should verify this isn't an
			// error
			selector, err := kube_meta.LabelSelectorAsSelector(ns.Selector)
			if err != nil {
				return Unknown, errors.Wrap(err, "internal error: couldn't convert to selector")
			}

			if !selector.Matches(kube_labels.Set(routeNs.GetLabels())) {
				listeners[l.Name] = NotPermitted
				continue
			}
		case gatewayapi.NamespacesFromSame:
			if gateway.Namespace != routeNs.GetName() {
				listeners[l.Name] = NotPermitted
				continue
			}
		case gatewayapi.NamespacesFromAll:
		}

		return Allowed, nil
	}

	sectionName := ""
	if refSectionName != nil {
		sectionName = string(*refSectionName)
	}

	// Look through the potential Listeners:
	// If it's our specific listener, we return that status
	for name, status := range listeners {
		if string(name) == sectionName {
			return status, nil
		}
	}

	// If we don't find our listener, our ref is invalid
	if sectionName != "" {
		return Invalid, nil
	}

	// If we aren't attaching to a specific listener then
	// as soon as one attaches we're attached (see ParentRef.Name docs)
	for _, status := range listeners {
		if status == Allowed {
			return Allowed, nil
		}
	}

	return NotPermitted, nil
}

// getParentRefGateway returns a kuma-class Gateway if one was found,
// otherwise it returns nil.
func getParentRefGateway(
	ctx context.Context,
	client kube_client.Client,
	fromNamespace string,
	ref gatewayapi.ParentRef,
) (*gatewayapi.Gateway, error) {
	name := string(ref.Name)
	// Group and Kind both have default values
	group := string(*ref.Group)
	kind := string(*ref.Kind)

	namespace := fromNamespace
	if ns := ref.Namespace; ns != nil {
		namespace = string(*ns)
	}

	if group != gatewayapi.GroupName || kind != "Gateway" {
		return nil, nil
	}

	gateway := &gatewayapi.Gateway{}

	if err := client.Get(ctx, kube_types.NamespacedName{Namespace: namespace, Name: name}, gateway); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	class, err := common.GetGatewayClass(ctx, client, gateway.Spec.GatewayClassName)
	if err != nil {
		return nil, err
	}

	if class.Spec.ControllerName != common.ControllerName {
		return nil, nil
	}

	return gateway, nil
}

// EvaluateParentRefAttachment reports whether a route in the given namespace can attach
// via the given ParentRef.
func EvaluateParentRefAttachment(
	ctx context.Context,
	client kube_client.Client,
	routeNs *kube_core.Namespace,
	ref gatewayapi.ParentRef,
) (Attachment, error) {
	gateway, err := getParentRefGateway(ctx, client, routeNs.GetName(), ref)
	if err != nil {
		return Unknown, errors.Wrap(err, "couldn't find Gateway referrent")
	}
	if gateway == nil {
		return Unknown, nil
	}

	return findRouteListenerAttachment(gateway, routeNs, ref.SectionName)
}
