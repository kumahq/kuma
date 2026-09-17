package labels

import (
	"fmt"

	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation"

	"github.com/kumahq/kuma/v3/pkg/core/validators"
	"github.com/kumahq/kuma/v3/pkg/util/maps"
)

// ValidateOwnership rejects control-plane-owned labels an untrusted writer supplied
// with a value other than the one Compute stores for this write, unless the value is
// the one already stored (Previous), which Compute replaces. Violations are keyed by
// label in registry order. A label whose Compute fails is left to Compute to reject.
func ValidateOwnership(w Write, cp ControlPlane) validators.ValidationError {
	var err validators.ValidationError
	if w.TrustedWriter {
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
		if previous, ok := w.Previous[d.Key]; ok && previous == supplied {
			continue
		}
		computed, ok, computeErr := d.Compute(w, cp)
		switch {
		case computeErr != nil:
		case !ok:
			err.AddViolationAt(validators.Root().Key(d.Key), fmt.Sprintf("label %q is managed by the control plane and cannot be set here", d.Key))
		case computed != supplied:
			err.AddViolationAt(validators.Root().Key(d.Key), fmt.Sprintf("label %q is managed by the control plane: got %q, expected %q", d.Key, supplied, computed))
		}
	}
	return err
}

// ValidateFormat rejects malformed label keys and values whoever the writer is: the
// per-label format rules first, then the generic syntax rules in key order.
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
		for _, msg := range d.ValidateFormat(v) {
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
