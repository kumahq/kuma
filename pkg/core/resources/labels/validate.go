package labels

import (
	"fmt"

	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
	"github.com/kumahq/kuma/v3/pkg/util/maps"
)

// ValidateOwnership rejects an untrusted update or delete of a resource another control
// plane owns, then control-plane-owned labels an untrusted writer supplied with a value
// other than the one Compute stores for this write. A value equal to the stored one is
// the object round-tripping through the writer and is skipped, unless Compute removes
// the label: accepting it would drop the label silently. A label whose Compute fails is
// skipped too: Compute rejects that write. Violations are keyed by label in registry order.
func ValidateOwnership(w Write, cp ControlPlane) validators.ValidationError {
	var err validators.ValidationError
	if w.TrustedWriter {
		return err
	}
	if (cp.Mode == config_core.Global || cp.FederatedZone) && !isLocal(w.Namespace, w.StoredLabels, cp) {
		owner := "the global"
		if cp.Mode == config_core.Global {
			owner = "a zone"
		}
		err.AddViolationAt(
			validators.Root().Key(mesh_proto.ResourceOriginLabel),
			fmt.Sprintf("the resource is owned by %s control plane and can be changed only there", owner),
		)
		return err
	}
	for _, d := range registry {
		if d.Owner != OwnerControlPlane {
			continue
		}
		supplied, ok := w.Labels[d.Key]
		if !ok {
			continue
		}
		computed, ok, computeErr := d.Compute(d.Key, w, cp)
		stored, isStored := w.StoredLabels[d.Key]
		switch {
		case computeErr != nil:
		case !ok:
			err.AddViolationAt(validators.Root().Key(d.Key), fmt.Sprintf("label %q is managed by the control plane and cannot be set here", d.Key))
		case isStored && stored == supplied:
		case computed != supplied:
			err.AddViolationAt(validators.Root().Key(d.Key), fmt.Sprintf("label %q is managed by the control plane: got %q, expected %q", d.Key, supplied, computed))
		}
	}
	return err
}

// ValidateFormat rejects malformed label keys and values whoever the writer is: the
// per-label format rules first, then the generic syntax rules in key order. Reserved
// keys this control plane does not know are rejected there too, for untrusted writers.
func ValidateFormat(w Write) validators.ValidationError {
	err := validateRegisteredFormat(w)
	err.Add(validateSyntax(w))
	return err
}

// Validate is the API server's one call. The order is the one its responses always
// had: the per-label format rules, then ownership, then the generic syntax rules.
func Validate(w Write, cp ControlPlane) validators.ValidationError {
	err := validateRegisteredFormat(w)
	err.Add(ValidateOwnership(w, cp))
	err.Add(validateSyntax(w))
	return err
}

func validateRegisteredFormat(w Write) validators.ValidationError {
	var err validators.ValidationError
	for _, d := range registry {
		if d.ValidateFormat == nil {
			continue
		}
		v, ok := w.Labels[d.Key]
		if !ok {
			continue
		}
		for _, msg := range d.ValidateFormat(d.Key, v) {
			err.AddViolationAt(validators.Root().Key(d.Key), msg)
		}
	}
	return err
}

func validateSyntax(w Write) validators.ValidationError {
	var err validators.ValidationError
	for _, k := range maps.SortedKeys(w.Labels) {
		v := w.Labels[k]
		for _, msg := range validation.IsQualifiedName(k) {
			err.AddViolationAt(validators.Root().Key(k), msg)
		}
		if _, known := lookup(k); !known && !w.TrustedWriter && mesh_proto.IsReservedLabelKey(k) {
			err.AddViolationAt(validators.Root().Key(k), fmt.Sprintf("label %q is reserved and not known to this control plane", k))
		}
		if storedAsAnnotation(k) {
			for _, msg := range apimachineryvalidation.NameIsDNSSubdomain(v, false) {
				err.AddViolationAt(validators.Root().Key(k), msg)
			}
			continue
		}
		for _, msg := range validation.IsValidLabelValue(v) {
			err.AddViolationAt(validators.Root().Key(k), msg)
		}
	}
	return err
}
