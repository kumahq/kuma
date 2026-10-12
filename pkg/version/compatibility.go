package version

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"

	"github.com/kumahq/kuma/v3/pkg/core"
)

var log = core.Log.WithName("version").WithName("compatibility")

var PreviewVersionPrefix = "preview"

func IsPreviewVersion(version string) bool {
	return strings.Contains(version, PreviewVersionPrefix)
}

// ServerVersionHigher returns true when the server version's major version is
// higher than the client version's. A server from a newer major version can
// expose fields the client does not know about, and the client silently drops
// them when reading or applying resources.
func ServerVersionHigher(clientVersionStr, serverVersionStr string) bool {
	clientMajor, ok := majorVersion(clientVersionStr)
	if !ok {
		return false
	}

	serverMajor, ok := majorVersion(serverVersionStr)
	if !ok {
		return false
	}

	return serverMajor > clientMajor
}

// majorVersion parses the major version out of a version string. Preview tags
// are not always strict semver, so fall back to the segment before the first
// dot. Versions we cannot parse at all are assumed to be dev versions.
func majorVersion(version string) (uint64, bool) {
	v, err := semver.NewVersion(version)
	if err == nil {
		// Non-release builds report a 0.0.0 version, which would compare lower than any real server version and warn spuriously. // EXC:FILE011:documents-the-dev-version-sentinel
		if v.Major() == 0 {
			return 0, false
		}
		return v.Major(), true
	}

	if i := strings.IndexByte(version, '.'); i > 0 {
		if major, err := strconv.ParseUint(version[:i], 10, 64); err == nil {
			return major, true
		}
	}

	log.Info("cannot parse semantic version", "version", version)
	return 0, false
}

// CompatibleMinors is how many minor releases apart two components may be
// while still being reported as compatible. The last minor of the previous
// major counts as the minor directly before X.0 of the next one, so the
// window spans a major bump like any other.
const CompatibleMinors = 2

// DeploymentVersionCompatible returns true if the given component version
// is compatible with the installed version of Kuma CP.
// For all binaries which share a common version (Kuma DP, CP, Zone CP...), we
// support backwards compatibility of at most CompatibleMinors prior minor
// versions, where the last minor of the previous major counts as the minor
// directly before the current major's X.0. Without that the only supported
// upgrade path onto a new major - the previous major's last release - would
// report every one of its data planes and zone control planes as incompatible
// for the whole rollout.
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

	minMinor := max(int64(kumaVersion.Minor())-CompatibleMinors, 0)

	maxMinor := kumaVersion.Minor() + CompatibleMinors

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
// minor of the major line directly preceding the other, and the other is still
// within CompatibleMinors of that major bump, so X.0 (one minor after the last
// minor of the previous major) and X.1 (two) both accept it. Older minors of
// the previous major are not carried over, and the window still closes two
// minors in, so 2.13.x <-> 3.0.x and 2.14.x <-> 3.2.x stay incompatible.
func previousMajorCompatible(a, b *semver.Version) bool {
	older, newer := a, b
	if newer.Major() < older.Major() {
		older, newer = newer, older
	}

	lastMinor, ok := lastMinorOfMajor[older.Major()]

	return ok &&
		older.Major()+1 == newer.Major() &&
		older.Minor() == lastMinor &&
		newer.Minor() < CompatibleMinors
}
