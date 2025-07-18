package attachment

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_labels "k8s.io/apimachinery/pkg/labels"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi/common"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

type Attachment int

const (
	// Unknown means we shouldn't report on whether it's attached or not
	Unknown Attachment = iota
	// NotPermitted means the route isn't allowed to attach to the ref
	NotPermitted
	NoHostnameIntersection
	NoMatchingParent
	// Allowed means we could successfully attach
	Allowed
)

type Kind int

const (
	UnknownKind Kind = iota
	Gateway
	Service
)

// checkListenerFrom reports whether this ref is allowed to attach to
// the Gateway or specific Listener.
func checkListenerFrom(
	gatewayNs string,
	routeNs kube_client.Object,
	attemptedAttachments []gatewayapi.Listener,
) (Attachment, error) {
	// If we don't have any attachments, our ref has NoMatchingParent
	if len(attemptedAttachments) == 0 {
		return NoMatchingParent, nil
	}

	// Build a map of whether attaching to each listener is possible
	listenerAttachments := map[gatewayapi.SectionName]Attachment{}

	for _, l := range attemptedAttachments {
		ns := l.AllowedRoutes.Namespaces

		// The gateway controller ensures Kinds is either empty or contains
		// HTTPRoute, so we can ignore it here.

		// From determines whether we are permitted to attach to this ParentRef
		switch *ns.From {
		case gatewayapi_v1.NamespacesFromSelector:
			// TODO, the gateway controller/webhook should verify this isn't an
			// error
			selector, err := kube_meta.LabelSelectorAsSelector(ns.Selector)
			if err != nil {
				return Unknown, errors.Wrap(err, "internal error: couldn't convert to selector")
			}

			if !selector.Matches(kube_labels.Set(routeNs.GetLabels())) {
				listenerAttachments[l.Name] = NotPermitted
			} else {
				listenerAttachments[l.Name] = Allowed
			}
		case gatewayapi_v1.NamespacesFromSame:
			if gatewayNs != routeNs.GetName() {
				listenerAttachments[l.Name] = NotPermitted
			} else {
				listenerAttachments[l.Name] = Allowed
			}
		case gatewayapi_v1.NamespacesFromAll:
			listenerAttachments[l.Name] = Allowed
		}
	}

	// If we aren't attaching to a specific listener then
	// as soon as one attaches we're attached (see ParentRef.Name docs)
	for _, listener := range attemptedAttachments {
		if listenerAttachments[listener.Name] == Allowed {
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
	ref gatewayapi.ParentReference,
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

	if class == nil || class.Spec.ControllerName != common.ControllerName {
		return nil, nil
	}

	return gateway, nil
}

func nameMatchesWildcardName(name, pat gatewayapi.Hostname) bool {
	suffix := strings.TrimPrefix(string(pat), "*.")
	if len(suffix) == len(pat) {
		return false
	}

	return strings.HasSuffix(string(name), suffix) && suffix != string(name)
}

func hostnamesIntersect(routeHostnames, listenerHostnames []gatewayapi.Hostname) bool {
	if len(routeHostnames) == 0 {
		return true
	}

	for _, routeHostname := range routeHostnames {
		for _, listenerHostname := range listenerHostnames {
			if listenerHostname == "" {
				return true
			}
			if routeHostname == listenerHostname {
				return true
			}
			// If one hostname is a wildcard and the other hostname matches the wildcard
			if nameMatchesWildcardName(routeHostname, listenerHostname) || nameMatchesWildcardName(listenerHostname, routeHostname) {
				return true
			}
		}
	}
	return false
}

func evaluateGatewayAttachment(
	ctx context.Context,
	client kube_client.Client,
	routeHostnames []gatewayapi.Hostname,
	routeNs *kube_core.Namespace,
	ref gatewayapi.ParentReference,
) (Attachment, error) {
	gateway, err := getParentRefGateway(ctx, client, routeNs.GetName(), ref)
	if err != nil {
		return Unknown, errors.Wrap(err, "couldn't find Gateway referrent")
	}
	if gateway == nil {
		return Unknown, nil
	}

	var attemptedAttachments []gatewayapi.Listener
	var listenerHostnames []gatewayapi.Hostname
	for _, listener := range gateway.Spec.Listeners {
		if ref.SectionName != nil && *ref.SectionName != listener.Name {
			continue
		}
		if ref.Port != nil && *ref.Port != listener.Port {
			continue
		}
		attemptedAttachments = append(attemptedAttachments, listener)
		listenerHostnames = append(
			listenerHostnames,
			pointer.DerefOr(listener.Hostname, gatewayapi.Hostname("")),
		)
	}

	if !hostnamesIntersect(routeHostnames, listenerHostnames) {
		return NoHostnameIntersection, nil
	}

	return checkListenerFrom(gateway.Namespace, routeNs, attemptedAttachments)
}

// EvaluateParentRefAttachment reports whether a route in the given namespace can attach
// via the given ParentRef.
func EvaluateParentRefAttachment(
	ctx context.Context,
	client kube_client.Client,
	routeHostnames []gatewayapi.Hostname,
	routeNs *kube_core.Namespace,
	ref gatewayapi.ParentReference,
) (Attachment, Kind, error) {
	switch {
	case *ref.Kind == "Gateway" && *ref.Group == gatewayapi.GroupName:
		attachment, err := evaluateGatewayAttachment(ctx, client, routeHostnames, routeNs, ref)
		return attachment, Gateway, err
	case *ref.Kind == "Service" && (*ref.Group == kube_core.GroupName || *ref.Group == gatewayapi.GroupName):
		// Attaching to a Service can only affect requests coming from Services
		// in the same Namespace or requests going to the Service.
		return Allowed, Service, nil
	}
	return Unknown, UnknownKind, nil
}
