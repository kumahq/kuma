package versions

import (
	"fmt"
	"os"
	"time"

	"github.com/Masterminds/semver/v3"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/v3/pkg/version"
)

const previewVersion = "preview"

type Version struct {
	Version       string `json:"version"`
	Lts           bool   `json:"lts,omitempty"`
	EndOfLifeDate string `json:"endOfLifeDate"`
	ReleaseDate   string `json:"releaseDate"`
	SemVer        *semver.Version
}

func ParseFromFile(path string) []Version {
	content, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	plainVersions := []Version{}
	if err := yaml.Unmarshal(content, &plainVersions); err != nil {
		panic(err)
	}

	var versions []Version
	for _, v := range plainVersions {
		if v.Version == previewVersion {
			continue
		}

		ver, err := semver.NewVersion(v.Version)
		if err != nil {
			panic(err)
		}
		v.SemVer = ver
		versions = append(versions, v)
	}

	return versions
}

func UpgradableVersions(versions []Version, currentVersion semver.Version) []string {
	if currentVersion.Major() == 0 && currentVersion.Minor() == 0 && currentVersion.Patch() == 0 {
		// Assume we are minor+1 of the last version in versions
		currentVersion = versions[len(versions)-1].SemVer.IncMinor()
	}
	var res []string
	for _, version := range versions {
		if version.ReleaseDate == "" {
			continue
		}
		if version.EndOfLifeDate != "" {
			eol, err := time.Parse(time.DateOnly, version.EndOfLifeDate)
			if err != nil {
				panic(err)
			}
			if version.Lts && eol.After(time.Now()) {
				res = append(res, version.SemVer.String())
				continue
			}
		}
		if version.SemVer.LessThan(&currentVersion) && upgradableTo(version.SemVer, currentVersion, versions) {
			res = append(res, version.SemVer.String())
		}
	}
	if len(res) == 0 {
		panic(fmt.Sprintf("couldn't find version 2 minors behind current: %s", currentVersion))
	}
	return res
}

// upgradableTo reports whether an upgrade from version to current is within
// the supported window: at most two minors behind within the same major, or
// the last minor of the previous major (taken from versions) when current is
// X.0 or X.1. This mirrors pkg/version.DeploymentVersionCompatible, which
// treats the last minor of the previous major as the minor directly before X.0.
func upgradableTo(version *semver.Version, current semver.Version, versions []Version) bool {
	if version.Major() == current.Major() {
		return version.Minor()+2 >= current.Minor()
	}
	if version.Major()+1 != current.Major() || current.Minor() >= 2 {
		return false
	}
	return version.Minor() == lastMinorOfMajor(versions, version.Major())
}

// lastMinorOfMajor returns the highest minor of the given major line in versions.
func lastMinorOfMajor(versions []Version, major uint64) uint64 {
	var last uint64
	for _, v := range versions {
		if v.SemVer.Major() == major && v.SemVer.Minor() > last {
			last = v.SemVer.Minor()
		}
	}
	return last
}

func UpgradableVersionsFromBuild(versions []Version) []string {
	v := semver.MustParse(version.Build.Version)
	return UpgradableVersions(versions, *v)
}

// OldestSupportedVersionFromBuild returns the oldest version still within the
// compatibility window for the current build, derived from versions.yml.
func OldestSupportedVersionFromBuild(vers []Version) string {
	return UpgradableVersionsFromBuild(vers)[0]
}

// IsVersionLessThan reports whether version is strictly less than threshold.
func IsVersionLessThan(version, threshold string) bool {
	return semver.MustParse(version).LessThan(semver.MustParse(threshold))
}
