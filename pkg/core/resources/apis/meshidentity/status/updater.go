package status

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/kri"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshidentity_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshidentity/providers"
	meshtrust_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshtrust/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/core/user"
	"github.com/kumahq/kuma/pkg/util/pointer"
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
				mesh := mesh.NewMeshResource()
				if err := i.roResManager.Get(ctx, mesh, store.GetByKey(mid.GetMeta().GetMesh(), core_model.NoMesh)); err != nil {
					i.logger.Error(err, "failed to list Meshes")
					continue
				}
				conditions := []common_api.Condition{}
				message := "Successfully initialized"
				generationConditionStatus := kube_meta.ConditionTrue
				reason := "Ready"
				initConditions := i.initialize(ctx, mid, mesh)
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

func (i *IdentityProviderReconciler) initialize(ctx context.Context, mid *meshidentity_api.MeshIdentityResource, mesh *mesh.MeshResource) []common_api.Condition {
	conditions := []common_api.Condition{}
	if mesh.Spec.MeshServicesMode() != mesh_proto.Mesh_MeshServices_Exclusive {
		conditions = append(conditions, common_api.Condition{
			Type:    meshidentity_api.DependenciesReadyType,
			Status:  kube_meta.ConditionFalse,
			Reason:  "MeshServicesDisabled",
			Message: "MeshIdentity requires MeshServices to be enabled on the mesh. To enable, set `spec.meshServices.mode: Exclusive` on the mesh.",
		})
		return conditions
	}
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
	if err := provider.Initialize(ctx, mid); err != nil {
		conditions = append(conditions, common_api.Condition{
			Type:    meshidentity_api.ProviderConditionType,
			Status:  kube_meta.ConditionFalse,
			Reason:  "ProviderInitializationError",
			Message: err.Error(),
		})
		return conditions
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
	if mid.Spec.Provider.Type == meshidentity_api.BundledType &&
		pointer.DerefOr(mid.Spec.Provider.Bundled.MeshTrustCreation, meshidentity_api.MeshTrustCreationEnabled) == meshidentity_api.MeshTrustCreationEnabled {
		if err := i.createOrUpdateMeshTrust(ctx, mid); err != nil {
			conditions = append(conditions, common_api.Condition{
				Type:    meshidentity_api.MeshTrustConditionType,
				Status:  kube_meta.ConditionFalse,
				Reason:  "MeshTrustCreationError",
				Message: err.Error(),
			})
			return conditions
		} else {
			conditions = append(conditions, common_api.Condition{
				Type:    meshidentity_api.MeshTrustConditionType,
				Status:  kube_meta.ConditionTrue,
				Reason:  "MeshTrustCreated",
				Message: "MeshTrust has been successfully created",
			})
		}
	}

	return conditions
}

func (i *IdentityProviderReconciler) createOrUpdateMeshTrust(ctx context.Context, identity *meshidentity_api.MeshIdentityResource) error {
	meshTrust := meshtrust_api.NewMeshTrustResource()
	meshName := identity.Meta.GetMesh()
	resourceName := identity.Meta.GetName()

	trustDomain, err := identity.Spec.GetTrustDomain(identity.GetMeta(), i.zone)
	if err != nil {
		return err
	}
	// Check if the MeshTrust resource already exists
	update := true
	if err := i.roResManager.Get(ctx, meshTrust, store.GetByKey(resourceName, meshName)); err != nil {
		if store.IsNotFound(err) {
			update = false
		} else {
			return err
		}
	}
	ca, err := i.loadCA(ctx, identity)
	if err != nil {
		return err
	}

	origin := kri.From(identity).String()
	caPEM := string(ca)

	if update {
		// Check if the CA PEM is already present in the MeshTrust resource
		for _, bundle := range meshTrust.Spec.CABundles {
			if pemEqual(bundle.PEM.Value, caPEM) {
				// Already exists; no need to update
				return nil
			}
		}

		// Append the new CA PEM
		meshTrust.Spec.CABundles = append(meshTrust.Spec.CABundles, meshtrust_api.CABundle{
			Type: meshtrust_api.PemCABundleType,
			PEM: &meshtrust_api.PEM{
				Value: caPEM,
			},
		})

		return i.resManager.Update(ctx, meshTrust)
	}

	// Resource doesn't exist, create a new one
	meshTrust.Spec = &meshtrust_api.MeshTrust{
		Origin: &meshtrust_api.Origin{
			KRI: pointer.To(origin),
		},
		TrustDomain: trustDomain,
		CABundles: []meshtrust_api.CABundle{
			{
				Type: meshtrust_api.PemCABundleType,
				PEM: &meshtrust_api.PEM{
					Value: caPEM,
				},
			},
		},
	}
	return i.resManager.Create(ctx, meshTrust, store.CreateByKey(resourceName, meshName))
}

func (i *IdentityProviderReconciler) loadCA(ctx context.Context, identity *meshidentity_api.MeshIdentityResource) ([]byte, error) {
	provider, found := i.providers[string(identity.Spec.Provider.Type)]
	if !found {
		return nil, fmt.Errorf("provider: %s not found", identity.Spec.Provider.Type)
	}
	return provider.GetRootCA(ctx, identity)
}

func (i *IdentityProviderReconciler) NeedLeaderElection() bool {
	return true
}

func pemEqual(pem1, pem2 string) bool {
	block1, _ := pem.Decode([]byte(pem1))
	block2, _ := pem.Decode([]byte(pem2))
	if block1 == nil || block2 == nil {
		return false
	}
	cert1, err1 := x509.ParseCertificate(block1.Bytes)
	cert2, err2 := x509.ParseCertificate(block2.Bytes)
	if err1 != nil || err2 != nil {
		return bytes.Equal(block1.Bytes, block2.Bytes)
	}
	return cert1.Equal(cert2)
}
