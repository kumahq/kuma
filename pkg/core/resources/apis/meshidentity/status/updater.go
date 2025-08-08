package status

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	meshidentity_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshidentity/providers"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/core/user"
)

const (
	ReadyType     string = "Ready"
	ProviderType  string = "Provider"
	MeshTrustType string = "MeshTrustCreated"
)

type IdentityProviderReconciler struct {
	roResManager      manager.ReadOnlyResourceManager
	resManager        manager.ResourceManager
	logger            logr.Logger
	reconcileInterval time.Duration
	providers         providers.IdentityProviders
	zone              string
}

var _ component.Component = &IdentityProviderReconciler{}

func New(
	logger logr.Logger,
	reconcileInterval time.Duration,
	resManager manager.ResourceManager,
	roResManager manager.ReadOnlyResourceManager,
	providers providers.IdentityProviders,
	zone string,
) (*IdentityProviderReconciler, error) {
	return &IdentityProviderReconciler{
		logger:            logger,
		reconcileInterval: reconcileInterval,
		resManager:        resManager,
		roResManager:      roResManager,
		providers:         providers,
		zone:              zone,
	}, nil
}

func (i *IdentityProviderReconciler) Start(stop <-chan struct{}) error {
	i.logger.Info("starting")
	ticker := time.NewTicker(i.reconcileInterval)
	ctx := user.Ctx(context.Background(), user.ControlPlane)
	for {
		select {
		case <-ticker.C:
			mids := &meshidentity_api.MeshIdentityResourceList{}
			if err := i.roResManager.List(ctx, mids, store.ListOrdered()); err != nil {
				i.logger.Error(err, "failed to list MeshIdentities")
				continue
			}
			for _, mid := range mids.Items {
				conditions := []common_api.Condition{}
				message := "Successfully initialized"
				generationConditionStatus := kube_meta.ConditionTrue
				reason := "Ready"
				initConditions := i.initialize(ctx, mid)
				conditions = append(conditions, initConditions...)
				for _, condition := range initConditions {
					if condition.Status == kube_meta.ConditionFalse {
						message = "One of initialization steps failed"
						generationConditionStatus = kube_meta.ConditionFalse
						reason = "Failure"
					}
				}

				conditions = append(conditions, common_api.Condition{
					Type:    meshidentity_api.ReadyConditionType,
					Status:  generationConditionStatus,
					Reason:  reason,
					Message: message,
				})

				if mid.Status == nil {
					mid.Status = &meshidentity_api.MeshIdentityStatus{}
				}
				needsUpdate := false
				if !reflect.DeepEqual(conditions, mid.Status.Conditions) {
					mid.Status.Conditions = conditions
					needsUpdate = true
				}
				if needsUpdate {
					if err := i.resManager.Update(ctx, mid); err != nil {
						i.logger.Error(err, "failed to update MeshIdentity status", "meshIdentity", mid.GetMeta().GetName())
						continue
					}
				}
			}
		case <-stop:
			i.logger.Info("stopping")
			return nil
		}
	}
}

func (i *IdentityProviderReconciler) initialize(ctx context.Context, mid *meshidentity_api.MeshIdentityResource) []common_api.Condition {
	conditions := []common_api.Condition{}
	provider, found := i.providers[string(mid.Spec.Provider.Type)]
	if !found {
		conditions = append(conditions, common_api.Condition{
			Type:    meshidentity_api.ProviderConditionType,
			Status:  kube_meta.ConditionFalse,
			Reason:  "ProviderNotFoundError",
			Message: fmt.Sprintf("provider: %s not found", mid.Spec.Provider.Type),
		})
		return conditions
	}
	if !mid.Status.IsInitialized() {
		if err := provider.Initialize(ctx, mid); err != nil {
			conditions = append(conditions, common_api.Condition{
				Type:    meshidentity_api.ProviderConditionType,
				Status:  kube_meta.ConditionFalse,
				Reason:  "ProviderInitializationError",
				Message: err.Error(),
			})
			return conditions
		}
	}
	if err := provider.Validate(ctx, mid); err != nil {
		conditions = append(conditions, common_api.Condition{
			Type:    meshidentity_api.ProviderConditionType,
			Status:  kube_meta.ConditionFalse,
			Reason:  "ProviderValidationError",
			Message: err.Error(),
		})
		return conditions
	} else {
		conditions = append(conditions, common_api.Condition{
			Type:    meshidentity_api.ProviderConditionType,
			Status:  kube_meta.ConditionTrue,
			Reason:  "ProviderInitialized",
			Message: "Provider successfully initialized",
		})
	}

	return conditions
}

func (i *IdentityProviderReconciler) NeedLeaderElection() bool {
	return true
}
