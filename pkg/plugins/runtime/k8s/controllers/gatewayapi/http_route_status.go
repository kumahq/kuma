package gatewayapi

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_apimeta "k8s.io/apimachinery/pkg/api/meta"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi/common"
)

func (r *HTTPRouteReconciler) updateStatus(ctx context.Context, route *gatewayapi.HTTPRoute, conditions ParentConditions) error {
	updated := route.DeepCopy()
	mergeHTTPRouteStatus(ctx, updated, conditions)

	if err := r.Client.Status().Patch(ctx, updated, kube_client.MergeFrom(route)); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return nil
		}
		return errors.Wrap(err, "unable to update status subresource")
	}

	return nil
}

// mergeHTTPRouteStatus updates the route status with the list of conditions for
// each parent ref by mutating the given HTTPRoute.
func mergeHTTPRouteStatus(ctx context.Context, route *gatewayapi.HTTPRoute, parentConditions ParentConditions) {
	var mergedStatuses []gatewayapi.RouteParentStatus
	var previousStatuses []gatewayapi.RouteParentStatus

	// partition statuses based on whether we control them
	for _, status := range route.Status.Parents {
		if status.ControllerName != common.ControllerName {
			mergedStatuses = append(mergedStatuses, status)
		} else {
			previousStatuses = append(previousStatuses, status)
		}
	}

	// for each new parentstatus, either add it to the list or update the
	// existing one
	for ref, conditions := range parentConditions {
		previousStatus := gatewayapi.RouteParentStatus{
			ParentRef:      ref,
			ControllerName: common.ControllerName,
		}

		// Look through previous statuses for one belonging to the same ref
		// go abusing pointers as option types makes it very painful
		for _, candidatePreviousStatus := range previousStatuses {
			if reflect.DeepEqual(candidatePreviousStatus.ParentRef, ref) {
				previousStatus = candidatePreviousStatus
			}
		}

		for _, condition := range conditions {
			condition.ObservedGeneration = route.GetGeneration()
			kube_apimeta.SetStatusCondition(&previousStatus.Conditions, condition)
		}

		mergedStatuses = append(mergedStatuses, previousStatus)
	}

	route.Status.Parents = mergedStatuses
}
