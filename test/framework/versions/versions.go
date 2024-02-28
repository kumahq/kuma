package versions

import (
	"fmt"
	"os"

	"github.com/Masterminds/semver/v3"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/pkg/version"
)

const previewVersion = "preview"

func ParseFromFile(path string) []*semver.Version {
	content, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	plainVersions := []struct {
		Version string `json:"version"`
	}{}
	if err := yaml.Unmarshal(content, &plainVersions); err != nil {
		panic(err)
	}

	var versions []*semver.Version
	for _, v := range plainVersions {
		if v.Version == previewVersion {
			continue
		}

		ver, err := semver.NewVersion(v.Version)
		if err != nil {
			panic(err)
		}
		versions = append(versions, ver)
	}

	return versions
}

func OldestUpgradableToBuildVersion(versions []*semver.Version) string {
	currentVersion := *semver.MustParse(version.Build.Version)
	return OldestUpgradableToVersion(versions, currentVersion)
}

func OldestUpgradableToVersion(versions []*semver.Version, currentVersion semver.Version) string {
	if currentVersion.Major() == 0 && currentVersion.Minor() == 0 && currentVersion.Patch() == 0 {
		// Assume we are minor+1 of the last version in versions
		currentVersion = versions[len(versions)-1].IncMinor()
	}

	for _, version := range versions {
		if version.Major() == currentVersion.Major() && version.Minor() == currentVersion.Minor()-2 {
			return version.String()
		}
	}
	panic(fmt.Sprintf("couldn't find version 2 minors behind current: %s", currentVersion))
}
