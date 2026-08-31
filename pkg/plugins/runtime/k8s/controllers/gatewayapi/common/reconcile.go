package common

import (
	"context"
	"fmt"
	"hash/fnv"
	"maps"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"

	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	k8s_model "github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s/native/pkg/model"
	k8s_registry "github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s/native/pkg/registry"
)

const ownerLabel = "gateways.kuma.io/gateway.networking.k8s.io-owner"

// OwnedObject describes an object owned by a gateway-api resource,
// including the namespace it should live in.
type OwnedObject struct {
	Namespace string
	Spec      core_model.ResourceSpec
	Labels    map[string]string
}

func hashNamespacedName(name kube_types.NamespacedName) string {
	hash := fnv.New32()
	hash.Write([]byte(name.Namespace))
	hash.Write([]byte(name.Name))
	// our hash is 8 characters and our label can be 63
	return fmt.Sprintf("%.54s-%x", fmt.Sprintf("%s_%s", name.Namespace, name.Name), hash.Sum(nil))
}

func OwnedPolicyName(owner kube_types.NamespacedName) string {
	return fmt.Sprintf("%s.%s", owner.Name, owner.Namespace)
}

// ReconcileLabelledObject manages a set of owned kuma objects based on
// labels with the owner key.
// ownerMesh can be empty if the ownedSpec is nil.
// ownedType tells us what type the owned object is.
// ownedSpec should be set to nil if the object shouldn't exist.
func ReconcileLabelledObject(
	ctx context.Context,
	logger logr.Logger,
	registry k8s_registry.TypeRegistry,
	client kube_client.Client,
	owner kube_types.NamespacedName,
	ownerMesh string,
	ownedType k8s_registry.ResourceType,
	owned map[string]OwnedObject,
) error {
	log := logger.WithValues("type", ownedType, "name", owner.Name, "namespace", owner.Namespace)
	// First we list which existing objects are owned by this owner.
	// We expect either 0 or 1 and depending on whether routeSpec is nil
	// we either create an object or update or delete the existing one.
	ownerLabelValue := hashNamespacedName(owner)
	labels := kube_client.MatchingLabels{
		ownerLabel: ownerLabelValue,
	}

	existingList, err := registry.NewList(ownedType)
	if err != nil {
		return errors.Wrapf(err, "could not create list of owned %T", ownedType)
	}

	if err := client.List(ctx, existingList, labels); err != nil {
		return err
	}

	// Delete unneeded objects
	existingObjs := map[string]k8s_model.KubernetesObject{}
	for _, existing := range existingList.GetItems() {
		desired, ok := owned[existing.GetName()]
		if !ok || existing.GetNamespace() != desired.Namespace {
			err := client.Delete(ctx, existing)
			switch {
			case kube_apierrs.IsNotFound(err):
				log.V(1).Info("object not found. Nothing to delete")
			case err == nil:
				log.Info("object deleted")
			default:
				return err
			}
			// We don't care about this anymore
			continue
		}
		existingObjs[existing.GetName()] = existing
	}

	// We need a mesh when creating objects
	if len(owned) > 0 && ownerMesh == "" {
		return fmt.Errorf("could not reconcile object, owner mesh must not be empty")
	}

	for ownedName, ownedObj := range owned {
		desiredLabels := map[string]string{}
		for k, v := range ownedObj.Labels {
			desiredLabels[k] = v
		}
		desiredLabels[ownerLabel] = ownerLabelValue

		// Update existing
		if existing, ok := existingObjs[ownedName]; ok {
			existingSpec, err := existing.GetSpec()
			if err != nil {
				return err
			}
			labelsChanged := !maps.Equal(existing.GetLabels(), desiredLabels)
			if core_model.Equal(existingSpec, ownedObj.Spec) && !labelsChanged {
				log.V(1).Info("object is the same. Nothing to update")
				continue
			}
			existing.SetLabels(desiredLabels)
			existing.SetSpec(ownedObj.Spec)

			if err := client.Update(ctx, existing); err != nil {
				return errors.Wrapf(err, "could not update owned %T", ownedType)
			}
			log.Info("object updated")

			continue
		}

		// Or create new
		newObj, err := registry.NewObject(ownedType)
		if err != nil {
			return errors.Wrapf(err, "could not get new %T from registry", ownedType)
		}

		newObj.SetObjectMeta(
			&kube_meta.ObjectMeta{
				Name:      ownedName,
				Namespace: ownedObj.Namespace,
				Labels:    desiredLabels,
			},
		)
		newObj.SetMesh(ownerMesh)
		newObj.SetSpec(ownedObj.Spec)

		if err := client.Create(ctx, newObj); err != nil {
			return errors.Wrapf(err, "could not create owned %T", ownedType)
		}
		logger.Info("object created")
	}

	return nil
}
