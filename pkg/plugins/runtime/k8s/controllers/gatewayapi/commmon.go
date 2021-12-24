package gatewayapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1alpha2"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

const controllerName = gatewayapi.GatewayController("gateways.kuma.io/controller")

func getGatewayClass(ctx context.Context, client kube_client.Client, name gatewayapi.ObjectName) (*gatewayapi.GatewayClass, error) {
	class := &gatewayapi.GatewayClass{}
	classObjectKey := kube_types.NamespacedName{Name: string(name)}

	if err := client.Get(ctx, classObjectKey, class); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return nil, nil
		}

		return nil, errors.Wrapf(err, "failed to get GatewayClass %s", classObjectKey)
	}

	if class.Spec.ControllerName != controllerName {
		return nil, nil
	}

	return class, nil
}

func serviceTagForGateway(name kube_types.NamespacedName) map[string]string {
	return map[string]string{
		mesh_proto.ServiceTag: fmt.Sprintf("%s_%s_gateway", name.Name, name.Namespace),
	}
}

func gatewayForServiceTag(tag string) (kube_types.NamespacedName, error) {
	splits := strings.SplitN(tag, "_", 3)
	if len(splits) < 3 || splits[2] != "gateway" {
		return kube_types.NamespacedName{}, fmt.Errorf("tag does not represent a Gateway")
	}

	return kube_types.NamespacedName{Name: splits[0], Namespace: splits[1]}, nil
}
