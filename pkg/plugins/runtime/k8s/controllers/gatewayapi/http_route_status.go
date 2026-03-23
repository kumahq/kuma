package gatewayapi

import (
	"cmp"
	"context"
	"fmt"
	"reflect"
	"slices"

	"github.com/pkg/errors"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_apimeta "k8s.io/apimachinery/pkg/api/meta"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/controllers/gatewayapi/common"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
)

func (r *HTTPRouteReconciler) updateStatus(ctx context.Context, route *gatewayapi.HTTPRoute, conditions ParentConditions) error {
	updated := route.DeepCopy()
	mergeHTTPRouteStatus(updated, conditions)

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
func mergeHTTPRouteStatus(route *gatewayapi.HTTPRoute, parentConditions ParentConditions) {
	// we cannot set a `nil` list
	mergedStatuses := []gatewayapi.RouteParentStatus{}
	var previousStatuses []gatewayapi.RouteParentStatus

	// Partition existing statuses: keep other controllers' statuses as-is,
	// collect our previous statuses separately so we can match them by ref
	// below when parentConditions may contain new refs not yet in the status.
	for _, status := range route.Status.Parents {
		if status.ControllerName != common.ControllerName {
			mergedStatuses = append(mergedStatuses, status)
		} else {
			previousStatuses = append(previousStatuses, status)
		}
	}

	// For each parent ref in parentConditions, find the matching previous
	// status (if any) or create a new one. This cannot be merged with the
	// loop above because parentConditions may contain refs that have no
	// existing status yet.
	var ownedStatuses []gatewayapi.RouteParentStatus
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

		ownedStatuses = append(ownedStatuses, previousStatus)
	}

	// Sort our controlled statuses by parent ref to ensure deterministic
	// ordering and avoid unnecessary status patches caused by Go map
	// iteration order.
	slices.SortFunc(ownedStatuses, func(a, b gatewayapi.RouteParentStatus) int {
		return cmp.Compare(parentRefSortKey(a.ParentRef), parentRefSortKey(b.ParentRef))
	})

	route.Status.Parents = append(mergedStatuses, ownedStatuses...)
}

func parentRefSortKey(ref gatewayapi.ParentReference) string {
	port := ""
	if ref.Port != nil {
		port = fmt.Sprint(*ref.Port)
	}
	return fmt.Sprintf("%s/%s/%s/%s/%s/%s",
		pointer.Deref(ref.Group),
		pointer.Deref(ref.Kind),
		pointer.Deref(ref.Namespace),
		ref.Name,
		pointer.Deref(ref.SectionName),
		port,
	)
}
