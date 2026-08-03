package envoy

import (
	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
)

// VersionCompatible returns true if the given version of
// envoy is compatible with this DP, false otherwise, expectedVersion is in Masterminds/semver/v3 format.
func VersionCompatible(expectedVersion string, envoyVersion string) (bool, error) {
	if expectedVersion == envoyVersion {
		return true, nil
	}
	expectedVersion = "~" + expectedVersion
	ver, err := semver.NewVersion(envoyVersion)
	if err != nil {
		return false, errors.Wrapf(err, "unable to parse envoy version %s", envoyVersion)
	}

	constraint, err := semver.NewConstraint(expectedVersion)
	if err != nil {
		// Programmer error
		return false, errors.Wrapf(err, "Invalid envoy compatibility constraint %s", expectedVersion)
	}

	return constraint.Check(ver), nil
}
