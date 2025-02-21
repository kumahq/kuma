package versions

import (
	"fmt"
	"os"
	"time"

	"github.com/Masterminds/semver/v3"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/pkg/version"
)

const previewVersion = "preview"

type Version struct {
	Version       string `json:"version"`
	Release       string `json:"release"`
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
				res = append(res, version.Release)
				continue
			}
		}
		if version.SemVer.LessThan(&currentVersion) && version.SemVer.Major() == currentVersion.Major() && version.SemVer.Minor() >= currentVersion.Minor()-2 {
			res = append(res, version.Release)
		}
	}
	if len(res) == 0 {
		panic(fmt.Sprintf("couldn't find version 2 minors behind current: %s", currentVersion))
	}
	return res
}

func UpgradableVersionsFromBuild(versions []Version) []string {
	v := semver.MustParse(version.Build.Version)
	return UpgradableVersions(versions, *v)
}
