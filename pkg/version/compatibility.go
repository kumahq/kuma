package version

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"

	"github.com/kumahq/kuma/pkg/core"
)

var log = core.Log.WithName("version").WithName("compatibility")

var PreviewVersionPrefix = "preview"

func IsPreviewVersion(version string) bool {
	return strings.Contains(version, PreviewVersionPrefix)
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
