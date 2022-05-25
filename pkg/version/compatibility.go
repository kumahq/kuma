package version

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"

	"github.com/kumahq/kuma/pkg/core"
)

var log = core.Log.WithName("version").WithName("compatibility")

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
		"~1.4.0": {
			Envoy: "~1.18.4",
		},
		"~1.5.0": {
			Envoy: "~1.21.1",
		},
		"~1.6.0": {
			Envoy: "~1.21.1",
		},
		"~1.7.0": {
			Envoy: "~1.22.0",
		},
		// This includes all dev versions branched from the first release
		// candidate (i.e. both master and release-1.x)
		// and all 1.x releases and RCs. See Masterminds/semver#21
		"~1.6.1-anyprerelease": {
			Envoy: "~1.21.1",
		},
		"~1.7.0-anyprerelease": {
			Envoy: "~1.22.0",
		},
	},
}

var DevVersionPrefix = "dev"

// DeploymentVersionCompatible returns true if the given component version
// is compatible with the installed version of Kuma CP.
// For all binaries which share a common version (Kuma DP, CP, Zone CP...), we
// support backwards compatibility of at most two prior minor versions.
func DeploymentVersionCompatible(kumaVersionStr string, componentVersionStr string) bool {
	if strings.Contains(kumaVersionStr, "dev") || strings.Contains(componentVersionStr, "dev") {
		return true
	}

	kumaVersion, err := semver.NewVersion(kumaVersionStr)
	if err != nil {
		// Assume some kind of dev version
		log.Info("cannot parse semantic version", "version", kumaVersionStr)
		return true
	}

	componentVersion, err := semver.NewVersion(componentVersionStr)
	if err != nil {
		// Assume some kind of dev version
		log.Info("cannot parse semantic version", "version", componentVersionStr)
		return true
	}

	minMinor := int64(kumaVersion.Minor()) - 2
	if minMinor < 0 {
		minMinor = 0
	}

	maxMinor := kumaVersion.Minor() + 2

	constraint, err := semver.NewConstraint(
		fmt.Sprintf(">= %d.%d, <= %d.%d", kumaVersion.Major(), minMinor, kumaVersion.Major(), maxMinor),
	)

	if err != nil {
		// Programmer error
		panic(err)
	}

	return constraint.Check(componentVersion)
}
