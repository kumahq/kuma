package labels

import (
	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation"

	"github.com/kumahq/kuma/v3/pkg/core/validators"
	"github.com/kumahq/kuma/v3/pkg/util/maps"
)

// ValidateOwnership rejects control-plane-owned labels an untrusted writer supplied
// with a value the control plane would not have chosen. Violations are keyed by label
// in registry order.
func ValidateOwnership(w Write, cp ControlPlane) validators.ValidationError {
	var err validators.ValidationError
	if w.TrustedWriter {
		return err
	}
	for _, d := range registry {
		if d.Owner != OwnerControlPlane || d.ValidateValue == nil {
			continue
		}
		v, ok := w.Labels[d.Key]
		if !ok {
			continue
		}
		for _, msg := range d.ValidateValue(v, w, cp) {
			err.AddViolationAt(validators.Root().Key(d.Key), msg)
		}
	}
	return err
}

// ValidateFormat rejects malformed label keys and values whoever the writer is: the
// per-label format rules first, then the generic syntax rules in key order.
func ValidateFormat(w Write) validators.ValidationError {
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

// Validate is ValidateOwnership followed by ValidateFormat.
func Validate(w Write, cp ControlPlane) validators.ValidationError {
	err := ValidateOwnership(w, cp)
	err.Add(ValidateFormat(w))
	return err
}
