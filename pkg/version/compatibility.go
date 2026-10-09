package version

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"

	"github.com/kumahq/kuma/v2/pkg/core"
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

// DeploymentVersionCompatible returns true if the given component version
// is compatible with the installed version of Kuma CP.
// For all binaries which share a common version (Kuma DP, CP, Zone CP...), we
// support backwards compatibility of at most two prior minor versions.
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

	minMinor := max(int64(kumaVersion.Minor())-2, 0)

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
