package common

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1alpha2"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

const (
	ControllerName = gatewayapi.GatewayController("gateways.kuma.io/controller")
	HTTPRouteKind  = gatewayapi.Kind("HTTPRoute")
)

func ServiceTagForGateway(name kube_types.NamespacedName) map[string]string {
	return map[string]string{
		mesh_proto.ServiceTag: fmt.Sprintf("%s_%s_gateway", name.Name, name.Namespace),
	}
}
func GetGatewayClass(ctx context.Context, client kube_client.Client, name gatewayapi.ObjectName) (*gatewayapi.GatewayClass, error) {
	class := &gatewayapi.GatewayClass{}
	classObjectKey := kube_types.NamespacedName{Name: string(name)}

	if err := client.Get(ctx, classObjectKey, class); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return nil, nil
		}

		return nil, errors.Wrapf(err, "failed to get GatewayClass %s", classObjectKey)
	}

	return class, nil
}

// ParentRefMatchesGateway checks whether a ref points to the given Gateway.
func ParentRefMatchesGateway(routeNamespace string, parentRef gatewayapi.ParentRef, gateway *gatewayapi.Gateway) bool {
	referencedNamespace := routeNamespace
	if parentRef.Namespace != nil {
		referencedNamespace = string(*parentRef.Namespace)
	}

	// We're looking at all HTTPRoutes, at some point one may
	// reference a non-Gateway object.
	// We don't care whether a specific listener is referenced
	return *parentRef.Group == gatewayapi.GroupName &&
		*parentRef.Kind == "Gateway" &&
		referencedNamespace == gateway.Namespace &&
		string(parentRef.Name) == gateway.Name
}
