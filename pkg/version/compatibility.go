package version

import (
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
)

type DataplaneCompatibility struct {
	Envoy string `json:"envoy"`
}

type Compatibility struct {
	KumaDP map[string]DataplaneCompatibility `json:"kumaDp"`
}

var CompatibilityMatrix = Compatibility{
	KumaDP: map[string]DataplaneCompatibility{
		"1.0.0": {
			Envoy: "1.16.0",
		},
		"1.0.1": {
			Envoy: "1.16.0",
		},
		"1.0.2": {
			Envoy: "1.16.1",
		},
		"1.0.3": {
			Envoy: "1.16.1",
		},
		"1.0.4": {
			Envoy: "1.16.1",
		},
		"1.0.5": {
			Envoy: "1.16.2",
		},
		"1.0.6": {
			Envoy: "1.16.2",
		},
		"1.0.7": {
			Envoy: "1.16.2",
		},
		"1.0.8": {
			Envoy: "1.16.2",
		},
		"~1.1.0": {
			Envoy: "~1.17.0",
		},
		"~1.2.0": {
			Envoy: "~1.18.0",
		},
		"~1.3.0": {
			Envoy: "~1.18.4",
		},
	},
}

// DataplaneConstraints returns which Envoy should be used with given version of Kuma.
// This information is later used in the GUI as a warning.
// Kuma ships with given Envoy version, but user can use their own Envoy version (especially on Universal)
// therefore we need to inform them that they are not using compatible version.
func (c Compatibility) DataplaneConstraints(version string) (*DataplaneCompatibility, error) {
	v, err := semver.NewVersion(version)
	if err != nil {
		return nil, errors.Wrapf(err, "could not build a constraint %s", version)
	}

	var matchedCompat []DataplaneCompatibility
	for constraintRaw, dpCompat := range c.KumaDP {
		constraint, err := semver.NewConstraint(constraintRaw)
		if err != nil {
			return nil, errors.Wrapf(err, "could not build a constraint %s", constraintRaw)
		}
		if constraint.Check(v) {
			matchedCompat = append(matchedCompat, dpCompat)
		}
	}

	if len(matchedCompat) == 0 {
		return nil, errors.Errorf("no constraints for version: %s found", version)
	}

	if len(matchedCompat) > 1 {
		var matched []string
		for _, c := range matchedCompat {
			matched = append(matched, c.Envoy)
		}
		return nil, errors.Errorf(
			"more than one constraint for version %s: %s",
			version,
			strings.Join(matched, ", "),
		)
	}
	return &matchedCompat[0], nil
}
