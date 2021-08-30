package version

import (
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
)

type KumaDPCompatibility struct {
	Envoy string `json:"envoy"`
}

type Compatibility struct {
	KumaDP map[string]KumaDPCompatibility `json:"kumaDp"`
}

var CompatibilityMatrix = Compatibility{
	KumaDP: map[string]KumaDPCompatibility{
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

func (c Compatibility) DP(version string) (*KumaDPCompatibility, error) {
	v, err := semver.NewVersion(version)
	if err != nil {
		return nil, errors.Wrapf(err, "could not build a constraint %s", version)
	}

	var matchedCompat []KumaDPCompatibility
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
			"more than one constraint for version: %s\n%s",
			version,
			strings.Join(matched, "\n"),
		)
	}
	return &matchedCompat[0], nil
}
