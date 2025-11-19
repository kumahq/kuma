package callbacks

import (
	"context"
	"strings"

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

	if md.Resource == nil {
		return nil
	}

	mesh := proxyKey.Mesh
	labels := md.Resource.GetMeta().GetLabels()

	log := workloadLabelLog.
		WithValues("mesh", mesh).
		WithValues("proxyKey", proxyKey).
		WithValues("streamID", streamID)

	meshIdentities := &meshidentity_api.MeshIdentityResourceList{}
	if err := v.rm.List(ctx, meshIdentities, store.ListByMesh(mesh)); err != nil {
		log.Error(err, "failed to list MeshIdentities")
		return errors.Wrap(err, "failed to list MeshIdentities")
	}

	matched, found := meshidentity_api.BestMatched(labels, meshIdentities.Items)
	if !found {
		return nil
	}

	if matched.Spec.SpiffeID == nil {
		return nil
	}

	pathTemplate := pointer.DerefOr(matched.Spec.SpiffeID.Path, "")
	if !usesWorkloadLabel(pathTemplate) {
		return nil
	}

	if _, ok := labels[metadata.KumaWorkload]; !ok {
		miName := matched.Meta.GetName()
		errMsg := errors.Errorf(
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

func (v *WorkloadLabelValidator) OnProxyDisconnected(ctx context.Context, streamID core_xds.StreamID, proxyKey core_model.ResourceKey) {
	// No-op
}

// usesWorkloadLabel checks if the SPIFFE ID path template contains the kuma.io/workload label reference.
func usesWorkloadLabel(pathTemplate string) bool {
	return strings.Contains(pathTemplate, `label "kuma.io/workload"`) ||
		strings.Contains(pathTemplate, `label 'kuma.io/workload'`) ||
		strings.Contains(pathTemplate, "label `kuma.io/workload`")
}
