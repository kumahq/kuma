package util

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	kube_apimeta "k8s.io/apimachinery/pkg/api/meta"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_schema "k8s.io/apimachinery/pkg/runtime/schema"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const controllerKey string = ".metadata.controller"

// CreateOrUpdateControlled either creates an object to be controlled by the
// given object or updates the existing one.
// This object must be indexed by calling IndexControllerOf.
func CreateOrUpdateControlled(
	ctx context.Context, client kube_client.Client, owner kube_meta.Object, objectList kube_client.ObjectList, mutate func(kube_client.Object) (kube_client.Object, error),
) (kube_client.Object, error) {
	if err := client.List(
		ctx, objectList, kube_client.InNamespace(owner.GetNamespace()), kube_client.MatchingFields{controllerKey: owner.GetName()},
	); err != nil {
		return nil, errors.Wrap(err, "unable to list objects")
	}

	items, err := kube_apimeta.ExtractList(objectList)
	if err != nil {
		return nil, errors.Wrap(err, "unable to extract list of runtime.Objects")
	}

	switch len(items) {
	case 0:
		obj, err := mutate(nil)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't mutate object for creation")
		}

		if err := kube_controllerutil.SetControllerReference(owner, obj, client.Scheme()); err != nil {
			return nil, errors.Wrap(err, "unable to set object's controller reference")
		}

		if err := client.Create(ctx, obj); err != nil {
			return nil, errors.Wrap(err, "couldn't create object")
		}
		return obj, nil
	case 1:
		item, ok := items[0].(kube_client.Object)
		if !ok {
			return nil, fmt.Errorf("expected runtime.Object to be client.Object")
		}

		obj, err := mutate(item)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't mutate object for update")
		}

		if err := kube_controllerutil.SetControllerReference(owner, obj, client.Scheme()); err != nil {
			return nil, errors.Wrap(err, "unable to set object's controller reference")
		}

		if err := client.Update(ctx, obj); err != nil {
			return nil, errors.Wrap(err, "couldn't update object")
		}
		return obj, nil
	default:
		return nil, fmt.Errorf("expected a maximum of one item controlled by %s/%s", owner.GetNamespace(), owner.GetName())
	}
}

// IndexControllerOf adds an index to objects of this GVK for use with
// CreateOrUpdateControlled.
func IndexControllerOf(mgr kube_ctrl.Manager, ownerGVK kube_schema.GroupVersionKind, objType kube_client.Object) error {
	ownerKind := ownerGVK.Kind
	ownerGV := ownerGVK.GroupVersion().String()

	return mgr.GetFieldIndexer().IndexField(context.Background(), objType, controllerKey, func(rawObj kube_client.Object) []string {
		owner := kube_meta.GetControllerOf(rawObj)
		if owner == nil || owner.APIVersion != ownerGV || owner.Kind != ownerKind {
			return nil
		}

		return []string{owner.Name}
	})
}
