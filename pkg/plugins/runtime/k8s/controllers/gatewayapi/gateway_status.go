package gatewayapi

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_apimeta "k8s.io/apimachinery/pkg/api/meta"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"

	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi/common"
)

func (r *GatewayReconciler) updateStatus(
	ctx context.Context,
	gateway *gatewayapi.Gateway,
	gatewayInstance *mesh_k8s.MeshGatewayInstance,
	listenerConditions ListenerConditions,
) error {
	updated := gateway.DeepCopy()

	attachedListeners, err := attachedRoutesForListeners(ctx, gateway, r.Client)
	if err != nil {
		return err
	}

	mergeGatewayStatus(updated, gatewayInstance, listenerConditions, attachedListeners)

	if err := r.Client.Status().Patch(ctx, updated, kube_client.MergeFrom(gateway)); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return nil
		}
		return errors.Wrap(err, "unable to patch status subresource")
	}

	return nil
}

func gatewayAddresses(instance *mesh_k8s.MeshGatewayInstance) []gatewayapi_v1.GatewayStatusAddress {
	if instance == nil {
		return nil
	}

	ipType := gatewayapi.IPAddressType
	hostnameType := gatewayapi.HostnameAddressType

	var addrs []gatewayapi_v1.GatewayStatusAddress

	if lb := instance.Status.LoadBalancer; lb != nil {
		for _, addr := range instance.Status.LoadBalancer.Ingress {
			if addr.IP != "" {
				addrs = append(addrs, gatewayapi_v1.GatewayStatusAddress{
					Type:  &ipType,
					Value: addr.IP,
				})
			}
			if addr.Hostname != "" {
				addrs = append(addrs, gatewayapi_v1.GatewayStatusAddress{
					Type:  &hostnameType,
					Value: addr.Hostname,
				})
			}
		}
	}

	return addrs
}

const everyListener = gatewayapi.SectionName("")

type attachedRoutes struct {
	// num is the number of allowed routes for a listener
	num int32
}

// AttachedRoutesForListeners tracks the relevant status for routes pointing to a
// listener.
type AttachedRoutesForListeners map[gatewayapi.SectionName]attachedRoutes

// attachedRoutesForListeners returns a function that calculates the
// conditions for routes attached to a Gateway.
func attachedRoutesForListeners(
	ctx context.Context,
	gateway *gatewayapi.Gateway,
	client kube_client.Client,
) (AttachedRoutesForListeners, error) {

	// only watched namespaces
	var routes gatewayapi.HTTPRouteList
	if err := client.List(ctx, &routes, kube_client.MatchingFields{
		gatewayOfRouteIndexField: kube_client.ObjectKeyFromObject(gateway).String(),
	}); err != nil {
		return nil, errors.Wrap(err, "unexpected error listing HTTPRoutes")
	}
	attachedRoutes := AttachedRoutesForListeners{}

	for i := range routes.Items {
		route := routes.Items[i]
		for _, parentRef := range route.Spec.ParentRefs {
			sectionName := everyListener
			if parentRef.SectionName != nil {
				sectionName = *parentRef.SectionName
			}

			for i := range route.Status.Parents {
				refStatus := route.Status.Parents[i]
				if reflect.DeepEqual(refStatus.ParentRef, parentRef) {
					attached := attachedRoutes[sectionName]

					if kube_apimeta.IsStatusConditionTrue(refStatus.Conditions, string(gatewayapi.RouteConditionAccepted)) {
						attached.num++
					}

					attachedRoutes[sectionName] = attached
				}
			}
		}
	}

	return attachedRoutes, nil
}

// mergeGatewayListenerStatuses takes the statuses of the attached Routes and
// the other calculated conditions for this listener and returns a
// ListenerStatus.
func mergeGatewayListenerStatuses(
	gateway *gatewayapi.Gateway,
	conditions ListenerConditions,
	attachedRouteStatuses AttachedRoutesForListeners,
) []gatewayapi.ListenerStatus {
	previousStatuses := map[gatewayapi.SectionName]gatewayapi.ListenerStatus{}

	for _, status := range gateway.Status.Listeners {
		previousStatuses[status.Name] = status
	}

	var statuses []gatewayapi.ListenerStatus

	// for each new parentstatus, either add it to the list or update the
	// existing one
	for name, conditions := range conditions {
		var listener gatewayapi.Listener
		for _, l := range gateway.Spec.Listeners {
			if l.Name == name {
				listener = l
			}
		}

		supportedKinds := []gatewayapi.RouteGroupKind{}
		if len(listener.AllowedRoutes.Kinds) == 0 {
			g := gatewayapi.Group(gatewayapi.GroupVersion.Group)
			supportedKinds = append(supportedKinds,
				gatewayapi.RouteGroupKind{Group: &g, Kind: common.HTTPRouteKind},
			)
		}
		for _, rgk := range listener.AllowedRoutes.Kinds {
			if string(*rgk.Group) == gatewayapi.GroupVersion.Group && rgk.Kind == common.HTTPRouteKind {
				supportedKinds = append(supportedKinds, rgk)
			}
		}

		previousStatus := gatewayapi.ListenerStatus{
			Name:           name,
			AttachedRoutes: 0,
		}

		if prev, ok := previousStatuses[name]; ok {
			previousStatus = prev
		}

		previousStatus.SupportedKinds = supportedKinds

		for _, condition := range conditions {
			condition.ObservedGeneration = gateway.GetGeneration()
			kube_apimeta.SetStatusCondition(&previousStatus.Conditions, condition)
		}

		// Check resolved status for routes for this listener and for
		// non-specific parents.
		previousStatus.AttachedRoutes = attachedRouteStatuses[name].num + attachedRouteStatuses[everyListener].num

		statuses = append(statuses, previousStatus)
	}

	return statuses
}

// mergeGatewayStatus updates the status by mutating the given Gateway.
func mergeGatewayStatus(
	gateway *gatewayapi.Gateway,
	instance *mesh_k8s.MeshGatewayInstance,
	listenerConditions ListenerConditions,
	attachedListeners AttachedRoutesForListeners,
) {
	gateway.Status.Addresses = gatewayAddresses(instance)

	gateway.Status.Listeners = mergeGatewayListenerStatuses(gateway, listenerConditions, attachedListeners)

	programmedStatus := kube_meta.ConditionTrue
	programmedReason := string(gatewayapi_v1.GatewayReasonProgrammed)

	for _, listener := range gateway.Status.Listeners {
		if !kube_apimeta.IsStatusConditionTrue(listener.Conditions, string(gatewayapi_v1.ListenerConditionProgrammed)) {
			programmedStatus = kube_meta.ConditionFalse
			programmedReason = string(gatewayapi_v1.GatewayReasonInvalid)
		}
	}

	conditions := []kube_meta.Condition{
		{
			Type:   string(gatewayapi_v1.GatewayConditionAccepted),
			Status: kube_meta.ConditionTrue,
			Reason: string(gatewayapi_v1.GatewayReasonAccepted),
		},
		{
			Type:   string(gatewayapi_v1.GatewayConditionProgrammed),
			Status: programmedStatus,
			Reason: programmedReason,
		},
	}

	for _, c := range conditions {
		c.ObservedGeneration = gateway.GetGeneration()
		kube_apimeta.SetStatusCondition(&gateway.Status.Conditions, c)
	}
}
