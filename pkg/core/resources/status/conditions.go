// Package status provides shared helpers for managing resource status
// conditions across MeshService / MeshExternalService / MeshMultiZoneService
// and other resources that expose a Conditions slice.
package status

import (
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/kri"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/core"
	meshexternalservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshmzservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/xds/envoy/tls"
)

// ConditionEquals reports whether conditions already contains an entry of the
// same Type as newCondition with matching Status/Reason/Message.
func ConditionEquals(conditions []common_api.Condition, newCondition common_api.Condition) bool {
	for _, c := range conditions {
		if c.Type == newCondition.Type {
			return c.Status == newCondition.Status &&
				c.Reason == newCondition.Reason &&
				c.Message == newCondition.Message
		}
	}
	return false
}

// UpdateConditions replaces the existing condition of the same Type or
// appends newCondition if no such entry exists.
func UpdateConditions(conditions []common_api.Condition, newCondition common_api.Condition) []common_api.Condition {
	for i, c := range conditions {
		if c.Type == newCondition.Type {
			conditions[i] = newCondition
			return conditions
		}
	}
	return append(conditions, newCondition)
}

// BuildSNICompliantCondition computes the SNICompliant condition for a
// resource identified by id with the given ports. It is True when SNI
// generation succeeds for every port, False otherwise (with the first
// validation error surfaced in Message).
func BuildSNICompliantCondition(id kri.Identifier, ports []core.Port) common_api.Condition {
	if len(ports) == 0 {
		// No ports — conservative: report compliant, there's nothing to
		// generate an SNI for. Producing False here would flag every
		// half-populated resource during initial sync.
		return common_api.Condition{
			Type:    common_api.SNICompliantCondition,
			Status:  kube_meta.ConditionTrue,
			Reason:  common_api.SNICompliantReason,
			Message: "No ports defined; SNI compliance is vacuously satisfied.",
		}
	}
	for _, port := range ports {
		portID := kri.WithSectionName(id, port.GetName())
		if err := tls.ValidateSNIForKRI(portID); err != nil {
			return common_api.Condition{
				Type:    common_api.SNICompliantCondition,
				Status:  kube_meta.ConditionFalse,
				Reason:  common_api.SNINotCompliantReason,
				Message: "Resource is reachable only from within its own zone — cross-zone routing requires a compliant SNI. " + err.Error(),
			}
		}
	}
	return common_api.Condition{
		Type:    common_api.SNICompliantCondition,
		Status:  kube_meta.ConditionTrue,
		Reason:  common_api.SNICompliantReason,
		Message: "All ports produce SNIs that satisfy DNS naming and length limits.",
	}
}

// IsSNICompliant returns true when the destination's status reports SNI
// compliance. It is the source of truth for xDS generators that need to
// decide whether to emit cross-zone (mesh-scoped) constructs for a
// destination: when this returns false, the resource cannot be reached via
// the new SNI format and only its local-zone path should be configured.
//
// A missing condition is treated as compliant. The status updater runs on the
// Zone CP on a separate loop, so during the first interval after a resource
// is created the condition is not yet set; defaulting to true matches the
// behavior an operator would expect for a fresh, otherwise-valid resource.
// A misconfigured resource will be marked False on the next status tick and
// the next xDS snapshot picks that up.
//
// Resource types other than MeshService / MeshExternalService /
// MeshMultiZoneService are also treated as compliant — they don't go through
// the SNIFromKRI path.
func IsSNICompliant(res any) bool {
	var conditions []common_api.Condition
	switch r := res.(type) {
	case *meshservice_api.MeshServiceResource:
		if r == nil {
			return true
		}
		conditions = r.Status.Conditions
	case *meshexternalservice_api.MeshExternalServiceResource:
		if r == nil {
			return true
		}
		conditions = r.Status.Conditions
	case *meshmzservice_api.MeshMultiZoneServiceResource:
		if r == nil {
			return true
		}
		conditions = r.Status.Conditions
	default:
		return true
	}
	for _, c := range conditions {
		if c.Type == common_api.SNICompliantCondition {
			return c.Status == kube_meta.ConditionTrue
		}
	}
	return true
}
