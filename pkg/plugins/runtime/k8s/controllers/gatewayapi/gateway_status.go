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

func (r *GatewayReconciler) updateStatus(ctx context.Context, gateway *gatewayapi.Gateway, gatewayInstance *mesh_k8s.GatewayInstance) error {
	updated := gateway.DeepCopy()
	mergeGatewayStatus(updated, gatewayInstance)

	if err := r.Client.Status().Patch(ctx, updated, kube_client.MergeFrom(gateway)); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return nil
		}
		return errors.Wrap(err, "unable to patch status subresource")
	}

	return nil
}

// mergeGatewayStatus updates the status by mutating the given Gateway.
func mergeGatewayStatus(gateway *gatewayapi.Gateway, instance *mesh_k8s.GatewayInstance) {
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

	gateway.Status.Addresses = addrs

	readinessStatus := kube_meta.ConditionFalse
	readinessReason := gatewayapi.GatewayReasonListenersNotReady

	// TODO(michaelbeaumont) it'd be nice to get more up to date info from the
	// kuma-dp instance to tell whether listeners are _really_ ready
	if len(addrs) == 0 {
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
