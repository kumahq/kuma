package version

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"

	"github.com/kumahq/kuma/v3/pkg/core"
)

var log = core.Log.WithName("version").WithName("compatibility")

var PreviewVersionPrefix = "preview"

func IsPreviewVersion(version string) bool {
	return strings.Contains(version, PreviewVersionPrefix)
}

// compatibleMinors is how many minor versions apart two components may be
// while still being reported as compatible.
const compatibleMinors = 2

// lastMinorOfMajor records the final minor release of each major line that
// has been superseded by a new major. That release is treated as the minor
// directly preceding the next major's X.0, so the compatibility window spans
// the major bump the same way it spans any other minor bump: 2.14.x is
// compatible with 3.0.x and 3.1.x, while 2.13.x <-> 3.0.x and
// 2.14.x <-> 3.2.x are not.
var lastMinorOfMajor = map[uint64]uint64{
	2: 14,
}

// DeploymentVersionCompatible returns true if the given component version
// is compatible with the installed version of Kuma CP.
// For all binaries which share a common version (Kuma DP, CP, Zone CP...), we
// support backwards compatibility of at most two prior minor versions, where
// the last minor of the previous major counts as the minor before X.0.
func DeploymentVersionCompatible(kumaVersionStr, componentVersionStr string) bool {
	if IsPreviewVersion(kumaVersionStr) || IsPreviewVersion(componentVersionStr) {
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

	minMinor := max(int64(kumaVersion.Minor())-compatibleMinors, 0)

	maxMinor := kumaVersion.Minor() + compatibleMinors

	constraint, err := semver.NewConstraint(
		fmt.Sprintf(">= %d.%d, <= %d.%d", kumaVersion.Major(), minMinor, kumaVersion.Major(), maxMinor),
	)
	if err != nil {
		// Programmer error
		panic(err)
	}

	if constraint.Check(componentVersion) {
		return true
	}

	return previousMajorCompatible(kumaVersion, componentVersion)
}

// previousMajorCompatible reports whether one of the versions is the last
// minor of the major line directly preceding the other, and the other is
// still within compatibleMinors of that major bump (X.0 is one minor after
// the last minor of the previous major, X.1 is two).
func previousMajorCompatible(a, b *semver.Version) bool {
	older, newer := a, b
	if newer.Major() < older.Major() {
		older, newer = newer, older
	}

	lastMinor, ok := lastMinorOfMajor[older.Major()]

	return ok &&
		older.Major()+1 == newer.Major() &&
		older.Minor() == lastMinor &&
		newer.Minor() < compatibleMinors
}
