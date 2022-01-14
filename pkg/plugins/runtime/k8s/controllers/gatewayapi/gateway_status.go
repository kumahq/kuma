package gatewayapi

import (
	"context"

	"github.com/pkg/errors"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_apimeta "k8s.io/apimachinery/pkg/api/meta"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1alpha2"

	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
)

func (r *GatewayReconciler) updateStatus(
	ctx context.Context,
	gateway *gatewayapi.Gateway,
	gatewayInstance *mesh_k8s.GatewayInstance,
	listenerConditions ListenerConditions,
) error {
	updated := gateway.DeepCopy()
	mergeGatewayStatus(updated, gatewayInstance, listenerConditions)

	if err := r.Client.Status().Patch(ctx, updated, kube_client.MergeFrom(gateway)); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return nil
		}
		return errors.Wrap(err, "unable to patch status subresource")
	}

	return nil
}

func gatewayAddresses(gateway *gatewayapi.Gateway, instance *mesh_k8s.GatewayInstance) []gatewayapi.GatewayAddress {
	ipType := gatewayapi.IPAddressType
	hostnameType := gatewayapi.HostnameAddressType

	var addrs []gatewayapi.GatewayAddress

	if lb := instance.Status.LoadBalancer; lb != nil {
		for _, addr := range instance.Status.LoadBalancer.Ingress {
			if addr.IP != "" {
				addrs = append(addrs, gatewayapi.GatewayAddress{
					Type:  &ipType,
					Value: addr.IP,
				})
			}
			if addr.Hostname != "" {
				addrs = append(addrs, gatewayapi.GatewayAddress{
					Type:  &hostnameType,
					Value: addr.Hostname,
				})
			}
		}
	}

	return addrs
}

func mergeGatewayListenerStatuses(gateway *gatewayapi.Gateway, conditions ListenerConditions) []gatewayapi.ListenerStatus {
	previousStatuses := map[gatewayapi.SectionName]gatewayapi.ListenerStatus{}

	for _, status := range gateway.Status.Listeners {
		previousStatuses[status.Name] = status
	}

	var statuses []gatewayapi.ListenerStatus

	// for each new parentstatus, either add it to the list or update the
	// existing one
	for name, conditions := range conditions {
		previousStatus := gatewayapi.ListenerStatus{
			Name: name,
			// TODO it's difficult to determine this number with Kuma, so we
			// leave it at 0
			AttachedRoutes: 0,
			SupportedKinds: []gatewayapi.RouteGroupKind{{Kind: httpRouteKind}},
		}

		if prev, ok := previousStatuses[name]; ok {
			previousStatus = prev
		}

		for _, condition := range conditions {
			condition.ObservedGeneration = gateway.GetGeneration()
			kube_apimeta.SetStatusCondition(&previousStatus.Conditions, condition)
		}

		statuses = append(statuses, previousStatus)
	}

	return statuses
}

// mergeGatewayStatus updates the status by mutating the given Gateway.
func mergeGatewayStatus(gateway *gatewayapi.Gateway, instance *mesh_k8s.GatewayInstance, listenerConditions ListenerConditions) {
	gateway.Status.Addresses = gatewayAddresses(gateway, instance)

	gateway.Status.Listeners = mergeGatewayListenerStatuses(gateway, listenerConditions)

	readinessStatus := kube_meta.ConditionFalse
	readinessReason := gatewayapi.GatewayReasonListenersNotReady

	// TODO(michaelbeaumont) it'd be nice to get more up to date info from the
	// kuma-dp instance to tell whether listeners are _really_ ready
	if len(gateway.Status.Addresses) == 0 {
		readinessStatus = kube_meta.ConditionFalse
		readinessReason = gatewayapi.GatewayReasonAddressNotAssigned
	} else if kube_apimeta.IsStatusConditionTrue(instance.Status.Conditions, mesh_k8s.GatewayInstanceReady) {
		readinessStatus = kube_meta.ConditionTrue
		readinessReason = gatewayapi.GatewayReasonReady
	}

	conditions := []kube_meta.Condition{
		{
			Type:   string(gatewayapi.GatewayConditionScheduled),
			Status: kube_meta.ConditionTrue,
			Reason: string(gatewayapi.GatewayReasonScheduled),
		},
		{
			Type:   string(gatewayapi.GatewayConditionReady),
			Status: readinessStatus,
			Reason: string(readinessReason),
		},
	}

	for _, c := range conditions {
		c.ObservedGeneration = gateway.GetGeneration()
		kube_apimeta.SetStatusCondition(&gateway.Status.Conditions, c)
	}
}
