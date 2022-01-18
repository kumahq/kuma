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

// ManageControlledObject is used to handle the lifecycle of and mutate
// a single object controlled by owner.
// This object type must be indexed by calling IndexControllerOf.
// If a controlled object exists, it's passed to mutate. Otherwise mutate
// receives a nil object.
// mutate should return a nil Object if the owned object should be deleted or
// not created.
func ManageControlledObject(
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

		// We don't want to create anything
		if obj == nil {
			return nil, nil
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

		original := item.DeepCopyObject().(kube_client.Object)

		obj, err := mutate(item)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't mutate object for update")
		}

		// We want to delete our object
		if obj == nil {
			err := client.Delete(ctx, item)
			return nil, errors.Wrap(err, "couldn't delete object")
		}

		if err := kube_controllerutil.SetControllerReference(owner, obj, client.Scheme()); err != nil {
			return nil, errors.Wrap(err, "unable to set object's controller reference")
		}

		if err := client.Patch(ctx, obj, kube_client.MergeFrom(original)); err != nil {
			return nil, errors.Wrap(err, "couldn't patch object")
		}
		return obj, nil
	default:
		return nil, fmt.Errorf("expected a maximum of one item controlled by %s/%s", owner.GetNamespace(), owner.GetName())
	}
}

// IndexControllerOf adds an index to objects of this GVK for use with
// ManageControlledObject.
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
