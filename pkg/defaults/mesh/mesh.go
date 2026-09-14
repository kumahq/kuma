package mesh

import (
	"context"
	"sync"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/v3/pkg/core"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/tokens"
	kuma_log "github.com/kumahq/kuma/v3/pkg/log"
)

var log = core.Log.WithName("defaults").WithName("mesh")

// ensureMux protects concurrent EnsureDefaultMeshResources invocation.
// On Kubernetes, EnsureDefaultMeshResources is called both from MeshManager when creating default Mesh and from the MeshController
// When they run concurrently:
// 1 invocation can check that a default resource is absent and then create it.
// 2 invocation can check that the same resource is absent, but it was just created, so it tries to create it which results in error
var ensureMux = sync.Mutex{}

func EnsureDefaultMeshResources(
	ctx context.Context,
	resManager manager.ResourceManager,
	mesh model.Resource,
	extensions context.Context,
) error {
	ensureMux.Lock()
	defer ensureMux.Unlock()

	meshName := mesh.GetMeta().GetName()
	logger := kuma_log.AddFieldsFromCtx(log, ctx, extensions).WithValues("mesh", meshName)

	logger.Info("ensuring default resources for Mesh exist")

	created, err := ensureDataplaneTokenSigningKey(ctx, resManager, mesh)
	if err != nil {
		return errors.Wrap(err, "could not create default Dataplane Token Signing Key")
	}
	if created {
		resKey := tokens.SigningKeyResourceKey(system.DataplaneTokenSigningKey(meshName), tokens.DefaultKeyID, meshName)
		logger.Info("default Dataplane Token Signing Key created", "name", resKey.Name)
	} else {
		logger.Info("Dataplane Token Signing Key already exists")
	}
	return nil
}
