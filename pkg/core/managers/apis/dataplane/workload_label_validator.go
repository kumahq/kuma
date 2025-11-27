package dataplane

import (
	"context"
	"fmt"

	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"

	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
)

type workloadLabelValidator struct{}

var _ Validator = &workloadLabelValidator{}

func NewWorkloadLabelValidator() Validator {
	return &workloadLabelValidator{}
}

func (w *workloadLabelValidator) ValidateCreate(_ context.Context, _ model.ResourceKey, newDp *core_mesh.DataplaneResource, _ *core_mesh.MeshResource) error {
	return w.validateWorkloadLabel(newDp)
}

func (w *workloadLabelValidator) ValidateUpdate(_ context.Context, newDp *core_mesh.DataplaneResource, _ *core_mesh.MeshResource) error {
	return w.validateWorkloadLabel(newDp)
}

func (w *workloadLabelValidator) validateWorkloadLabel(dp *core_mesh.DataplaneResource) error {
	labels := dp.GetMeta().GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}

	workloadName, ok := labels[metadata.KumaWorkload]
	if !ok || workloadName == "" {
		return nil
	}

	allErrs := apimachineryvalidation.NameIsDNS1035Label(workloadName, false)
	if len(allErrs) != 0 {
		return fmt.Errorf("invalid %s label value %q: must be a valid DNS-1035 label (at most 63 characters, matching regex [a-z]([-a-z0-9]*[a-z0-9])?): %v",
			metadata.KumaWorkload, workloadName, allErrs)
	}

	return nil
}
