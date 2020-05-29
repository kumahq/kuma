package server

import (
	"context"
	"reflect"

	"github.com/pkg/errors"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/store"
)

var (
	ingressLog = core.Log.WithName("ingress-reconcile")
)

func NewIngressReconciler(resourceManager manager.ResourceManager) *IngressReconciler {
	return &IngressReconciler{
		resourceManager: resourceManager,
		lastUsedPort:    10000,
	}
}

type IngressReconciler struct {
	resourceManager manager.ResourceManager
	lastUsedPort    uint32
}

func (r *IngressReconciler) Reconcile(dataplaneId core_model.ResourceKey) error {
	ctx := context.Background()
	var dp core_mesh.DataplaneResource
	if err := r.resourceManager.Get(ctx, &dp, store.GetBy(dataplaneId)); err != nil {
		if store.IsResourceNotFound(err) {
			ingressLog.V(1).Info("Dataplane not found.", "dataplaneId", dataplaneId)
			return nil
		}
		return err
	}

	if dp.Spec.GetNetworking().GetIngress() == nil {
		return errors.Errorf("reconciliation loop works only for Ingress Dataplane")
	}

	meshes, err := r.getMeshes(ctx)
	if err != nil {
		return err
	}
	others, err := r.getDataplanes(ctx, meshes)
	if err != nil {
		return err
	}

	dp.Spec.Networking.Inbound = r.generateInbounds(others, dp.Spec.Networking.Inbound)
	if err := r.resourceManager.Update(ctx, &dp); err != nil {
		return err
	}
	return nil
}

type inboundSet []*mesh_proto.Dataplane_Networking_Inbound

func (set inboundSet) getBy(tags map[string]string) *mesh_proto.Dataplane_Networking_Inbound {
	for _, in := range set {
		if reflect.DeepEqual(in.Tags, tags) {
			return in
		}
	}
	return nil
}

func (r *IngressReconciler) generatePort() uint32 {
	r.lastUsedPort++
	return r.lastUsedPort
}

func (r *IngressReconciler) generateInbounds(others []*core_mesh.DataplaneResource, old []*mesh_proto.Dataplane_Networking_Inbound) []*mesh_proto.Dataplane_Networking_Inbound {
	inbounds := make([]*mesh_proto.Dataplane_Networking_Inbound, 0, len(others))
	for _, dp := range others {
		if dp.Spec.GetNetworking().GetIngress() != nil {
			continue
		}
		for _, dpInbound := range dp.Spec.GetNetworking().GetInbound() {
			if dup := inboundSet(inbounds).getBy(dpInbound.GetTags()); dup != nil {
				continue
			}
			var port uint32
			if prev := inboundSet(old).getBy(dpInbound.GetTags()); prev != nil {
				port = prev.Port
			} else {
				port = r.generatePort()
			}

			inbounds = append(inbounds, &mesh_proto.Dataplane_Networking_Inbound{
				Port: port, //  picked automatically
				Tags: dpInbound.GetTags(),
			})
		}
	}
	return inbounds
}

func (r *IngressReconciler) getMeshes(ctx context.Context) ([]*core_mesh.MeshResource, error) {
	meshList := &core_mesh.MeshResourceList{}
	if err := r.resourceManager.List(ctx, meshList); err != nil {
		return nil, err
	}

	return meshList.Items, nil
}

func (r *IngressReconciler) getDataplanes(ctx context.Context, meshes []*core_mesh.MeshResource) ([]*core_mesh.DataplaneResource, error) {
	dataplanes := make([]*core_mesh.DataplaneResource, 0)
	for _, mesh := range meshes {
		dataplaneList := &core_mesh.DataplaneResourceList{}
		if err := r.resourceManager.List(ctx, dataplaneList, store.ListByMesh(mesh.Meta.GetName())); err != nil {
			return nil, err
		}
		dataplanes = append(dataplanes, dataplaneList.Items...)
	}
	return dataplanes, nil
}
