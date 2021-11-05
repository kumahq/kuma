package gatewayapi

import (
	kube_apps "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func conditionOn(
	obj kube_client.Object, typ gatewayapi.GatewayConditionType, status metav1.ConditionStatus, reason gatewayapi.GatewayConditionReason,
) metav1.Condition {
	return metav1.Condition{
		Type: string(typ), Status: status, Reason: string(reason), LastTransitionTime: metav1.Now(), ObservedGeneration: obj.GetGeneration(),
	}
}

func getCondition(deployment *kube_apps.Deployment, typ kube_apps.DeploymentConditionType) *metav1.ConditionStatus {
	for _, c := range deployment.Status.Conditions {
		if c.Type == typ {
			status := metav1.ConditionStatus(c.Status)
			return &status
		}
	}

	return nil
}

func setConditions(gateway *gatewayapi.Gateway, deployment *kube_apps.Deployment) {
	conditions := []metav1.Condition{
		conditionOn(gateway, gatewayapi.GatewayConditionScheduled, metav1.ConditionTrue, gatewayapi.GatewayReasonScheduled),
	}

	// TODO(michaelbeaumont) it'd be nice to get more up to date info from the
	// kuma-dp instance to tell whether listeners are _really_ ready
	if len(gateway.Status.Addresses) == 0 {
		conditions = append(conditions,
			conditionOn(gateway, gatewayapi.GatewayConditionReady, metav1.ConditionFalse, gatewayapi.GatewayReasonAddressNotAssigned),
		)
	} else if condition := getCondition(deployment, kube_apps.DeploymentAvailable); condition == nil || *condition != metav1.ConditionTrue {
		conditions = append(conditions,
			conditionOn(gateway, gatewayapi.GatewayConditionReady, metav1.ConditionFalse, gatewayapi.GatewayReasonListenersNotReady),
		)
	} else {
		conditions = append(conditions,
			conditionOn(gateway, gatewayapi.GatewayConditionReady, metav1.ConditionTrue, gatewayapi.GatewayReasonReady),
		)
	}

	gateway.Status.Conditions = conditions
}
