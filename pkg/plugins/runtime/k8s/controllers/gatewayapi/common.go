package gatewayapi

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1alpha2"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	k8s_model "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

const (
	controllerName = gatewayapi.GatewayController("gateways.kuma.io/controller")
	httpRouteKind  = gatewayapi.Kind("HTTPRoute")
)

func getGatewayClass(ctx context.Context, client kube_client.Client, name gatewayapi.ObjectName) (*gatewayapi.GatewayClass, error) {
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

func serviceTagForGateway(name kube_types.NamespacedName) map[string]string {
	return map[string]string{
		mesh_proto.ServiceTag: fmt.Sprintf("%s_%s_gateway", name.Name, name.Namespace),
	}
}

// reconcileLabelledObjectSet manages a set of owned kuma objects based on
// labels with the owner key.
// ownedType tells us what type the owned object is.
// ownedSpec should be set to nil if the object shouldn't exist.
func reconcileLabelledObject(
	ctx context.Context,
	registry k8s_registry.TypeRegistry,
	client kube_client.Client,
	owner kube_types.NamespacedName,
	ownedType k8s_registry.ResourceType,
	ownedSpec proto.Message,
) error {
	// First we list which existing objects are owned by this owner.
	// We expect either 0 or 1 and depending on whether routeSpec is nil
	// we either create an object or update or delete the existing one.
	ownerLabelValue := fmt.Sprintf("%s-%s", owner.Namespace, owner.Name)
	labels := kube_client.MatchingLabels{
		ownerLabel: ownerLabelValue,
	}

	ownedList, err := registry.NewList(ownedType)
	if err != nil {
		return errors.Wrapf(err, "could not create list of owned %T", ownedType)
	}

	if err := client.List(ctx, ownedList, labels); err != nil {
		return err
	}

	if l := len(ownedList.GetItems()); l > 1 {
		return fmt.Errorf("internal error: found %d items labeled as owned by this object, expected either zero or one", l)
	}

	var existing k8s_model.KubernetesObject
	if items := ownedList.GetItems(); len(items) == 1 {
		existing = items[0]
	}

	if ownedSpec == nil {
		if existing != nil {
			if err := client.Delete(ctx, existing); err != nil && !kube_apierrs.IsNotFound(err) {
				return err
			}
		}
		return nil
	}

	if existing != nil {
		existing.SetSpec(ownedSpec)

		if err := client.Update(ctx, existing); err != nil {
			return errors.Wrapf(err, "could not update owned %T", ownedType)
		}
		return nil
	}

	owned, err := registry.NewObject(ownedType)
	if err != nil {
		return errors.Wrapf(err, "could not get new %T from registry", ownedType)
	}

	owned.SetObjectMeta(
		&kube_meta.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", ownerLabelValue),
			Labels: map[string]string{
				ownerLabel: ownerLabelValue,
			},
		},
	)
	owned.SetSpec(ownedSpec)

	if err := client.Create(ctx, owned); err != nil {
		return errors.Wrapf(err, "could not create owned %T", ownedType)
	}

	return nil
}
