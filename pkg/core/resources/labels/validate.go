package labels

import (
	"fmt"

	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
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
		for _, msg := range d.ValidateValue(d.Key, v, w, cp) {
			err.AddViolationAt(validators.Root().Key(d.Key), msg)
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

// ValidateUpdate rejects labels an untrusted update would change although the
// control plane keeps them fixed for the life of the object. Previous must hold the
// stored labels and Labels the ones the update stores.
func ValidateUpdate(w Write, cp ControlPlane) validators.ValidationError {
	var err validators.ValidationError
	if w.TrustedWriter {
		return err
	}
	for _, d := range registry {
		if d.ValidateUpdate == nil {
			continue
		}
		previous, ok := w.Previous[d.Key]
		if !ok {
			continue
		}
		for _, msg := range d.ValidateUpdate(previous, w.Labels[d.Key], w, cp) {
			err.AddViolationAt(validators.Root().Key(d.Key), msg)
		}
	}
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
