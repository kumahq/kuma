package gatewayapi

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1alpha2"

	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
)

func conditionOn(
	obj kube_client.Object, typ gatewayapi.GatewayConditionType, status metav1.ConditionStatus, reason gatewayapi.GatewayConditionReason,
) metav1.Condition {
	return metav1.Condition{
		Type: string(typ), Status: status, Reason: string(reason), LastTransitionTime: metav1.Now(), ObservedGeneration: obj.GetGeneration(),
	}
}

func getReadyCondition(instance *mesh_k8s.GatewayInstance) *metav1.ConditionStatus {
	for _, c := range instance.Status.Conditions {
		if c.Type == mesh_k8s.GatewayInstanceReady {
			status := c.Status
			return &status
		}
	}

	return nil
}

func setConditions(gateway *gatewayapi.Gateway, instance *mesh_k8s.GatewayInstance) {
	readinessStatus := metav1.ConditionUnknown                      //nolint:ineffassign
	readinessReason := gatewayapi.GatewayConditionReason("Unknown") //nolint:ineffassign

	// TODO(michaelbeaumont) it'd be nice to get more up to date info from the
	// kuma-dp instance to tell whether listeners are _really_ ready
	if len(gateway.Status.Addresses) == 0 {
		readinessStatus = metav1.ConditionFalse
		readinessReason = gatewayapi.GatewayReasonAddressNotAssigned
	} else if ready := getReadyCondition(instance); ready != nil && *ready == metav1.ConditionTrue {
		readinessStatus = metav1.ConditionTrue
		readinessReason = gatewayapi.GatewayReasonReady
	} else {
		readinessStatus = metav1.ConditionFalse
		readinessReason = gatewayapi.GatewayReasonListenersNotReady
	}

	gateway.Status.Conditions = []metav1.Condition{
		conditionOn(gateway, gatewayapi.GatewayConditionScheduled, metav1.ConditionTrue, gatewayapi.GatewayReasonScheduled),
		conditionOn(gateway, gatewayapi.GatewayConditionReady, readinessStatus, readinessReason),
	}
}
