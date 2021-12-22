package gatewayapi

import (
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1alpha2"

	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
)

func gatewayCondition(
	obj kube_client.Object,
	typ gatewayapi.GatewayConditionType,
	status kube_meta.ConditionStatus,
	reason gatewayapi.GatewayConditionReason,
) kube_meta.Condition {
	return kube_meta.Condition{
		Type:               string(typ),
		Status:             status,
		Reason:             string(reason),
		LastTransitionTime: kube_meta.Now(),
		ObservedGeneration: obj.GetGeneration(),
	}
}

func getReadyCondition(instance *mesh_k8s.GatewayInstance) *kube_meta.ConditionStatus {
	for _, c := range instance.Status.Conditions {
		if c.Type == mesh_k8s.GatewayInstanceReady {
			status := c.Status
			return &status
		}
	}

	return nil
}

func setConditions(gateway *gatewayapi.Gateway, instance *mesh_k8s.GatewayInstance) {
	readinessStatus := kube_meta.ConditionUnknown                   //nolint:ineffassign
	readinessReason := gatewayapi.GatewayConditionReason("Unknown") //nolint:ineffassign

	// TODO(michaelbeaumont) it'd be nice to get more up to date info from the
	// kuma-dp instance to tell whether listeners are _really_ ready
	if len(gateway.Status.Addresses) == 0 {
		readinessStatus = kube_meta.ConditionFalse
		readinessReason = gatewayapi.GatewayReasonAddressNotAssigned
	} else if ready := getReadyCondition(instance); ready != nil && *ready == kube_meta.ConditionTrue {
		readinessStatus = kube_meta.ConditionTrue
		readinessReason = gatewayapi.GatewayReasonReady
	} else {
		readinessStatus = kube_meta.ConditionFalse
		readinessReason = gatewayapi.GatewayReasonListenersNotReady
	}

	gateway.Status.Conditions = []kube_meta.Condition{
		gatewayCondition(gateway, gatewayapi.GatewayConditionScheduled, kube_meta.ConditionTrue, gatewayapi.GatewayReasonScheduled),
		gatewayCondition(gateway, gatewayapi.GatewayConditionReady, readinessStatus, readinessReason),
	}
}
