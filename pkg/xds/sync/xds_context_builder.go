package sync

import (
	"context"

	"github.com/kumahq/kuma/pkg/envoy/admin"

	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

type xdsContextBuilder struct {
	resManager            manager.ReadOnlyResourceManager
	connectionInfoTracker ConnectionInfoTracker
	lookupIP              lookup.LookupIPFunc
	envoyAdmin            admin.EnvoyAdmin

	cpContext *xds_context.ControlPlaneContext
}

func newXDSContextBuilder(
	cpContext *xds_context.ControlPlaneContext,
	connectionInfoTracker ConnectionInfoTracker,
	resManager manager.ReadOnlyResourceManager,
	lookupIP lookup.LookupIPFunc,
	envoyAdmin admin.EnvoyAdmin,
) *xdsContextBuilder {
	return &xdsContextBuilder{
		resManager:            resManager,
		connectionInfoTracker: connectionInfoTracker,
		lookupIP:              lookupIP,
		envoyAdmin:            envoyAdmin,
		cpContext:             cpContext,
	}
}

func (c *xdsContextBuilder) buildMeshedContext(streamId int64, meshName string, meshHash string) (*xds_context.Context, error) {
	ctx := context.Background()
	xdsCtx := c.buildContext(streamId)

	mesh := core_mesh.NewMeshResource()
	if err := c.resManager.Get(ctx, mesh, core_store.GetByKey(meshName, core_model.NoMesh)); err != nil {
		return nil, err
	}

	dataplanes, err := xds_topology.GetDataplanes(syncLog, context.Background(), c.resManager, c.lookupIP, meshName)
	if err != nil {
		return nil, err
	}
	xdsCtx.Mesh = xds_context.MeshContext{
		Resource:   mesh,
		Dataplanes: dataplanes,
		Hash:       meshHash,
	}
	return xdsCtx, nil
}

func (c *xdsContextBuilder) buildContext(streamId int64) *xds_context.Context {
	return &xds_context.Context{
		ControlPlane:   c.cpContext,
		ConnectionInfo: c.connectionInfoTracker.ConnectionInfo(streamId),
		Mesh:           xds_context.MeshContext{},
		EnvoyAdmin:     c.envoyAdmin,
	}
}
