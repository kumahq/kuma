package common

import (
	"context"
	"fmt"
	"hash/fnv"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	k8s_model "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

const ownerLabel = "gateways.kuma.io/gateway.networking.k8s.io-owner"

func hashNamespacedName(name kube_types.NamespacedName) string {
	hash := fnv.New32()
	hash.Write([]byte(name.Namespace))
	hash.Write([]byte(name.Name))
	// our hash is 8 characters and our label can be 63
	return fmt.Sprintf("%.54s-%x", fmt.Sprintf("%s_%s", name.Namespace, name.Name), hash.Sum(nil))
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
	ownedSpec proto.Message,
) error {
	log := logger.WithValues("type", ownedType, "name", owner.Name, "namespace", owner.Namespace)
	// First we list which existing objects are owned by this owner.
	// We expect either 0 or 1 and depending on whether routeSpec is nil
	// we either create an object or update or delete the existing one.
	ownerLabelValue := hashNamespacedName(owner)
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
			err := client.Delete(ctx, existing)
			switch {
			case kube_apierrs.IsNotFound(err):
				log.V(1).Info("object not found. Nothing to delete")
			case err == nil:
				log.Info("object deleted")
			default:
				return err
			}
		}
		return nil
	}

	// We need a mesh when creating the object
	if ownerMesh == "" {
		return fmt.Errorf("could not reconcile object, owner mesh must not be empty")
	}

	if existing != nil {
		existingSpec, err := existing.GetSpec()
		if err != nil {
			return err
		}
		if core_model.Equal(existingSpec, ownedSpec) {
			log.V(1).Info("object is the same. Nothing to update")
			return nil
		}
		existing.SetSpec(ownedSpec)

		if err := client.Update(ctx, existing); err != nil {
			return errors.Wrapf(err, "could not update owned %T", ownedType)
		}
		log.Info("object updated")
		return nil
	}

	owned, err := registry.NewObject(ownedType)
	if err != nil {
		return errors.Wrapf(err, "could not get new %T from registry", ownedType)
	}

	owned.SetMesh(ownerMesh)

	owned.SetObjectMeta(
		&kube_meta.ObjectMeta{
			Name: fmt.Sprintf("%s.%s", owner.Name, owner.Namespace),
			Labels: map[string]string{
				ownerLabel: ownerLabelValue,
			},
		},
	)
	owned.SetSpec(ownedSpec)

	if err := client.Create(ctx, owned); err != nil {
		return errors.Wrapf(err, "could not create owned %T", ownedType)
	}
	logger.Info("object created")

	return nil
}
