package envoy

import (
	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
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
		return false, errors.Wrapf(err, "unable to parse envoy version %s", envoyVersion)
	}

	constraint, err := semver.NewConstraint(EnvoyCompatibility)
	if err != nil {
		// Programmer error
		panic(errors.Wrapf(err, "Invalid envoy compatibility constraint %s", EnvoyCompatibility))
	}

	return constraint.Check(ver), nil
}
