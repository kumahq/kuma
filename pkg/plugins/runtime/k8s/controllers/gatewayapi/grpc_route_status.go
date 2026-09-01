package gatewayapi

import (
	"context"

	"github.com/pkg/errors"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1"
)

func (r *GRPCRouteReconciler) updateGRPCRouteStatus(ctx context.Context, route *gatewayapi.GRPCRoute, conditions ParentConditions) error {
	updated := route.DeepCopy()
	updated.Status.Parents = mergeRouteParentStatus(route.GetGeneration(), updated.Status.Parents, conditions)

	if err := r.Client.Status().Patch(ctx, updated, kube_client.MergeFrom(route)); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return nil
		}
		return errors.Wrap(err, "unable to update status subresource")
	}

	return nil
}
