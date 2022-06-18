package envoy

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

// EnvoyCompatibility is the description of envoy versions
// which are currently compatible with this data plane,
// in Masterminds/semver/v3 format.
var EnvoyCompatibility = "~1.21.1"

// EnvoyVersionCompatible returns true if the given version of
// envoy is compatible with this DP, false otherwise.
func EnvoyVersionCompatible(envoyVersion string) (bool, error) {
	ver, err := semver.NewVersion(envoyVersion)
	if err != nil {
		return false, fmt.Errorf("unable to parse envoy version %s: %w", envoyVersion, err)
	}

	constraint, err := semver.NewConstraint(EnvoyCompatibility)
	if err != nil {
		// Programmer error
		panic(fmt.Errorf("Invalid envoy compatibility constraint %s: %w", EnvoyCompatibility, err))
	}

	return constraint.Check(ver), nil
}
