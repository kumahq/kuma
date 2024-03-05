package gatewayapi

import (
	kube_apimeta "k8s.io/apimachinery/pkg/api/meta"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"
)

func prepareConditions(conditions []kube_meta.Condition) []kube_meta.Condition {
	for _, condition := range []kube_meta.Condition{
		{
			Type:   string(gatewayapi.RouteConditionAccepted),
			Status: kube_meta.ConditionTrue,
			Reason: string(gatewayapi.RouteReasonAccepted),
		}, {
			Type:   string(gatewayapi.RouteConditionResolvedRefs),
			Status: kube_meta.ConditionTrue,
			Reason: string(gatewayapi.RouteReasonResolvedRefs),
		},
	} {
		if kube_apimeta.FindStatusCondition(conditions, condition.Type) == nil {
			kube_apimeta.SetStatusCondition(&conditions, condition)
		}
	}

	return conditions
}
