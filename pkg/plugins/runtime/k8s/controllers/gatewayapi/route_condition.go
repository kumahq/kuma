package gatewayapi

import (
	"fmt"

	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func routeCondition(
	route kube_client.Object, typ gatewayapi.RouteConditionType, status kube_meta.ConditionStatus, reason string, message ...string,
) kube_meta.Condition {
	c := kube_meta.Condition{
		Type: string(typ), Status: status, Reason: reason, LastTransitionTime: kube_meta.Now(), ObservedGeneration: route.GetGeneration(),
	}
	if len(message) > 0 {
		var args []interface{}
		for _, arg := range message[1:] {
			args = append(args, interface{}(arg))
		}
		c.Message = fmt.Sprintf(message[0], args...)
	}

	return c
}
