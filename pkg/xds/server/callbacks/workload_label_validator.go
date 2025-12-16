package callbacks

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core"
	meshidentity_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
)

var workloadLabelLog = core.Log.WithName("xds").WithName("workload-label-validator")

// WorkloadLabelValidator validates that dataplanes have the required kuma.io/workload label
// when they are selected by a MeshIdentity that uses the workload label in its SPIFFE ID path template.
type WorkloadLabelValidator struct {
	rm manager.ReadOnlyResourceManager
}

var _ DataplaneCallbacks = &WorkloadLabelValidator{}

func NewWorkloadLabelValidator(rm manager.ReadOnlyResourceManager) *WorkloadLabelValidator {
	return &WorkloadLabelValidator{
		rm: rm,
	}
}

func (v *WorkloadLabelValidator) OnProxyConnected(
	streamID core_xds.StreamID,
	proxyKey core_model.ResourceKey,
	ctx context.Context,
	md core_xds.DataplaneMetadata,
) error {
	if md.GetProxyType() != mesh_proto.DataplaneProxyType {
		return nil
	}

	dp := md.GetDataplaneResource()
	if dp == nil {
		return nil
	}

	// Skip validation for gateway dataplanes
	if dp.Spec.IsBuiltinGateway() {
		return nil
	}

	mesh := proxyKey.Mesh
	labels := dp.GetMeta().GetLabels()

	log := workloadLabelLog.
		WithValues("mesh", mesh).
		WithValues("proxyKey", proxyKey).
		WithValues("streamID", streamID)

	// List MeshIdentities for this mesh. This operation is performed on every proxy connection,
	// but the ReadOnlyResourceManager provides cached results, so performance impact is minimal.
	meshIdentities := &meshidentity_api.MeshIdentityResourceList{}
	if err := v.rm.List(ctx, meshIdentities, store.ListByMesh(mesh)); err != nil {
		return errors.Wrap(err, "failed to list MeshIdentities")
	}

	matched, found := meshidentity_api.BestMatched(labels, meshIdentities.Items)
	if !found {
		return nil
	}

	if !matched.Spec.UsesWorkloadLabel() {
		return nil
	}

	if _, ok := labels[metadata.KumaWorkload]; !ok {
		pathTemplate := pointer.Deref(matched.Spec.SpiffeID.Path)
		miName := matched.Meta.GetName()
		errMsg := fmt.Errorf(
			"missing required label '%s' - dataplane is selected by MeshIdentity '%s' with path template '%s'",
			metadata.KumaWorkload,
			miName,
			pathTemplate,
		)
		log.Error(errMsg, "dataplane rejected - missing required workload label",
			"meshIdentity", miName,
			"pathTemplate", pathTemplate,
			"dataplaneLabels", labels,
		)
		return errMsg
	}

	return nil
}

func (v *WorkloadLabelValidator) OnProxyDisconnected(_ context.Context, _ core_xds.StreamID, _ core_model.ResourceKey) {
	// No-op
}
