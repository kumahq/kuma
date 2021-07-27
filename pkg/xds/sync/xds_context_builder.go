package sync

import (
	"context"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/envoy/admin"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

type xdsContextBuilder struct {
	resManager            manager.ReadOnlyResourceManager
	connectionInfoTracker ConnectionInfoTracker
	lookupIP              lookup.LookupIPFunc
	envoyAdminClient      admin.EnvoyAdminClient

	cpContext *xds_context.ControlPlaneContext
}

func newXDSContextBuilder(
	cpContext *xds_context.ControlPlaneContext,
	connectionInfoTracker ConnectionInfoTracker,
	resManager manager.ReadOnlyResourceManager,
	lookupIP lookup.LookupIPFunc,
	envoyAdminClient admin.EnvoyAdminClient,
) *xdsContextBuilder {
	return &xdsContextBuilder{
		resManager:            resManager,
		connectionInfoTracker: connectionInfoTracker,
		lookupIP:              lookupIP,
		envoyAdminClient:      envoyAdminClient,
		cpContext:             cpContext,
	}
}

func (c *xdsContextBuilder) buildMeshedContext(dpKey core_model.ResourceKey, meshHash string) (*xds_context.Context, error) {
	ctx := context.Background()
	xdsCtx, err := c.buildContext(dpKey)
	if err != nil {
		return nil, err
	}

	mesh := core_mesh.NewMeshResource()
	if err := c.resManager.Get(ctx, mesh, core_store.GetByKey(dpKey.Mesh, core_model.NoMesh)); err != nil {
		return nil, err
	}

	dataplanes, err := xds_topology.GetDataplanes(syncLog, context.Background(), c.resManager, c.lookupIP, dpKey.Mesh)
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

func (c *xdsContextBuilder) buildContext(dpKey core_model.ResourceKey) (*xds_context.Context, error) {
	connectionInfo := c.connectionInfoTracker.ConnectionInfo(dpKey)
	if connectionInfo == nil {
		return nil, errors.New("connection info is not found")
	}
	return &xds_context.Context{
		ControlPlane:     c.cpContext,
		ConnectionInfo:   *connectionInfo,
		Mesh:             xds_context.MeshContext{},
		EnvoyAdminClient: c.envoyAdminClient,
	}, nil
}
